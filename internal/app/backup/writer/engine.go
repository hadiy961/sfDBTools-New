package writer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/crypto"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/compress"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/errorlog"
	"sfdbtools/internal/ui/progress"
)

func summarizeStderr(stderr string, maxLines int, maxChars int) string {
	if stderr == "" {
		return ""
	}
	lines := strings.Split(stderr, "\n")
	if maxLines > 0 && len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	s := strings.TrimSpace(strings.Join(lines, "\n"))
	if maxChars > 0 && len(s) > maxChars {
		s = s[:maxChars] + "..."
	}
	return s
}

type Engine struct {
	Log      applog.Logger
	ErrorLog *errorlog.ErrorLogger
	Options  *types_backup.BackupDBOptions
}

func New(log applog.Logger, errLog *errorlog.ErrorLogger, opts *types_backup.BackupDBOptions) *Engine {
	return &Engine{Log: log, ErrorLog: errLog, Options: opts}
}

func (e *Engine) resolveEncryptionKeyIfNeeded() (string, error) {
	if !e.Options.Encryption.Enabled {
		return "", nil
	}

	resolvedKey, source, err := crypto.ResolveKey(
		e.Options.Encryption.Key,
		consts.ENV_BACKUP_ENCRYPTION_KEY,
		false, // tidak perlu prompt untuk backup
	)
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan kunci enkripsi: %w", err)
	}

	e.Log.Debugf("Kunci enkripsi didapat dari: %s", source)
	return resolvedKey, nil
}

func (e *Engine) createBufferedOutputFile(outputPath string) (*os.File, *bufio.Writer, error) {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal membuat file output: %w", err)
	}
	bufWriter := bufio.NewWriterSize(outputFile, consts.BackupWriterBufferSize)
	return outputFile, bufWriter, nil
}

func (e *Engine) createWriterPipeline(baseWriter io.Writer, compressionRequired bool, compressionType string, encryptionKey string) (io.Writer, []io.Closer, error) {
	var writer io.Writer = baseWriter
	var closers []io.Closer

	// Layer 1: Encryption (paling dekat dengan file)
	if e.Options.Encryption.Enabled {
		encryptingWriter, err := crypto.NewStreamEncryptor(writer, []byte(encryptionKey))
		if err != nil {
			return nil, nil, fmt.Errorf("gagal membuat encrypting writer: %w", err)
		}
		closers = append(closers, encryptingWriter)
		writer = encryptingWriter
	}

	// Layer 2: Compression
	if compressionRequired {
		compressionConfig := compress.CompressionConfig{
			Type:  compress.CompressionType(compressionType),
			Level: compress.CompressionLevel(e.Options.Compression.Level),
		}
		compressingWriter, err := compress.NewCompressingWriter(writer, compressionConfig)
		if err != nil {
			return nil, nil, fmt.Errorf("gagal membuat compressing writer: %w", err)
		}
		closers = append(closers, compressingWriter)
		writer = compressingWriter
	}

	return writer, closers, nil
}

func (e *Engine) isFatalMysqldumpError(err error, stderrOutput string) bool {
	if err == nil {
		return false
	}

	if stderrOutput == "" {
		e.Log.Debug("mysqldump error with empty stderr, treating as fatal")
		return true
	}

	stderrLower := strings.ToLower(stderrOutput)

	fatalPatterns := []string{
		"access denied",
		"unknown database",
		"unknown server",
		"can't connect",
		"connection refused",
		"got error:",
		"error:",
		"failed",
	}
	for _, p := range fatalPatterns {
		if strings.Contains(stderrLower, p) {
			return true
		}
	}

	nonFatalPatterns := []string{
		"couldn't read keys from table",
		"references invalid table(s) or column(s) or function(s)",
		"definer/invoker of view lack rights",
		"warning:",
	}
	for _, p := range nonFatalPatterns {
		if strings.Contains(stderrLower, p) {
			return false
		}
	}

	return true
}

// ExecuteMysqldumpWithPipe runs mysqldump and streams into file via (optional) compression and encryption.
func (e *Engine) ExecuteMysqldumpWithPipe(ctx context.Context, mysqldumpArgs []string, outputPath string, compressionRequired bool, compressionType string) (*types_backup.BackupWriteResult, error) {
	encryptionKey, err := e.resolveEncryptionKeyIfNeeded()
	if err != nil {
		return nil, err
	}

	spin := progress.NewSpinnerWithElapsed("Memproses backup database")
	spin.Start()
	defer spin.Stop()

	outputFile, bufWriter, err := e.createBufferedOutputFile(outputPath)
	if err != nil {
		return nil, err
	}
	defer outputFile.Close()
	defer func() {
		if flushErr := bufWriter.Flush(); flushErr != nil {
			e.Log.Errorf("Error flushing buffer: %v", flushErr)
		}
	}()

	writer, closers, err := e.createWriterPipeline(bufWriter, compressionRequired, compressionType, encryptionKey)
	if err != nil {
		return nil, err
	}
	defer func() {
		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				e.Log.Errorf("Error closing writer: %v", err)
			}
		}
	}()

	cmd := exec.CommandContext(ctx, "mysqldump", mysqldumpArgs...)

	monitor := newDatabaseMonitorWriter(writer, spin, e.Log)
	cmd.Stdout = monitor

	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	runErr := cmd.Run()

	monitor.Finish(runErr == nil)

	if runErr != nil {
		stderrOutput := stderrBuf.String()

		if e.ErrorLog != nil {
			_ = e.ErrorLog.LogWithOutput(map[string]interface{}{
				"type": "mysqldump_backup",
				"file": outputPath,
			}, stderrOutput, runErr)
		}

		if e.isFatalMysqldumpError(runErr, stderrOutput) {
			result := &types_backup.BackupWriteResult{
				StderrOutput: stderrOutput,
				FileSize:     0,
			}
			excerpt := summarizeStderr(stderrOutput, 20, 2000)
			if excerpt != "" {
				return result, fmt.Errorf("mysqldump gagal: %w (stderr: %s)", runErr, excerpt)
			}
			return result, fmt.Errorf("mysqldump gagal: %w", runErr)
		}
		excerpt := summarizeStderr(stderrOutput, 12, 1200)
		if excerpt != "" {
			e.Log.Warnf("mysqldump exit with non-fatal error, treated as warning. stderr (excerpt):\n%s", excerpt)
		} else {
			e.Log.Warn("mysqldump exit with non-fatal error, treated as warning")
		}
	}

	fileInfo, statErr := os.Stat(outputPath)
	var fileSize int64
	if statErr == nil {
		fileSize = fileInfo.Size()
	}

	result := &types_backup.BackupWriteResult{
		StderrOutput: stderrBuf.String(),
		FileSize:     fileSize,
	}

	return result, nil
}
