// File : internal/app/backup/writer/engine.go
// Deskripsi : Core backup execution engine dengan streaming pipeline
// Author : Hadiyatna Muflihun
// Last Modified : 20 Januari 2026

package writer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/crypto"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/compress"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/errorlog"
	"sfdbtools/internal/shared/execx"
	"sfdbtools/internal/ui/progress"
)

func summarizeStderr(stderr string, maxLines int, maxChars int) string {
	if stderr == "" {
		return ""
	}
	// Prioritas: mariadb-dump, fallback: mysqldump.
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

func (e *Engine) createBufferedOutputFile(outputPath string, permissions string) (*os.File, *bufio.Writer, error) {
	perm := parseFilePermissions(permissions, e.Log)
	outputFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
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

// isFatalDumpError menentukan apakah error dari mysqldump/mariadb-dump adalah fatal atau non-fatal.
// Strategi:
// 1. Primary: gunakan exit code sebagai indikator utama
//   - Exit 0: Success (tidak error)
//   - Exit 1: Ambiguous (bisa warning atau error ringan) - cek pattern
//   - Exit 2+: Fatal error
//
// 2. Secondary: pattern matching untuk edge cases (hanya untuk exit code 1)
// 3. Log stderr yang tidak ter-classify untuk future improvement
func (e *Engine) isFatalDumpError(err error, stderrOutput string, exitCode int) bool {
	if err == nil {
		return false
	}

	// Primary check: exit code
	if exitCode == 0 {
		return false
	}

	if exitCode >= 2 {
		e.Log.Debugf("mysqldump exit code %d (>=2), treating as fatal", exitCode)
		return true
	}

	// Exit code 1: ambiguous case, perlu secondary check
	if stderrOutput == "" {
		// No stderr tapi exit 1 = suspicious, treat as fatal
		e.Log.Debug("dump error with exit 1 and empty stderr, treating as fatal")
		return true
	}

	// Pattern matching untuk mendeteksi known non-fatal warnings pada exit code 1
	stderrLower := strings.ToLower(stderrOutput)

	// Known non-fatal patterns (biasanya warnings yang aman diabaikan)
	nonFatalPatterns := []string{
		"couldn't read keys from table",
		"references invalid table(s) or column(s) or function(s)",
		"definer/invoker of view lack rights",
		"warning:",
		"note:",
	}
	for _, p := range nonFatalPatterns {
		if strings.Contains(stderrLower, p) {
			e.Log.Debugf("mysqldump exit 1 matched non-fatal pattern: %s", p)
			return false
		}
	}

	// Known fatal patterns (untuk memastikan edge cases terdeteksi)
	fatalPatterns := []string{
		"access denied",
		"unknown database",
		"unknown server",
		"can't connect",
		"connection refused",
		"permission denied",
		"no such file or directory",
	}
	for _, p := range fatalPatterns {
		if strings.Contains(stderrLower, p) {
			e.Log.Debugf("mysqldump exit 1 matched fatal pattern: %s", p)
			return true
		}
	}

	// Unclassified exit code 1 dengan stderr yang tidak dikenali
	// Default ke fatal untuk safety, tapi log untuk future improvement
	excerpt := summarizeStderr(stderrOutput, 5, 500)
	e.Log.Warnf("Unclassified mysqldump exit 1 stderr (treating as fatal for safety): %s", excerpt)

	return true
}

// parseFilePermissions mengkonversi string permissions (e.g., "0600") ke os.FileMode
// Jika parsing gagal atau permissions kosong, return default 0600 (lebih restrictive)
func parseFilePermissions(permStr string, logger applog.Logger) os.FileMode {
	const defaultPerm = 0600

	if permStr == "" {
		return defaultPerm
	}

	// Parse octal string to uint32
	perm, err := strconv.ParseUint(permStr, 8, 32)
	if err != nil {
		logger.Warnf("Invalid file_permissions '%s', using default 0600: %v", permStr, err)
		return defaultPerm
	}

	return os.FileMode(perm)
}

// ExecuteMysqldumpWithPipe menjalankan dump command dan streaming ke file via (opsional) kompresi dan enkripsi.
// Prioritas: mariadb-dump, fallback: mysqldump.
func (e *Engine) ExecuteMysqldumpWithPipe(ctx context.Context, mysqldumpArgs []string, outputPath string, compressionRequired bool, compressionType string, permissions string) (*types_backup.BackupWriteResult, error) {
	encryptionKey, err := e.resolveEncryptionKeyIfNeeded()
	if err != nil {
		return nil, err
	}

	dumpBin, err := execx.ResolveMariaDBDumpOrMysqldump()
	if err != nil {
		return nil, err
	}

	spin := progress.NewSpinnerWithElapsed("Memproses backup database")
	spin.Start()
	defer spin.Stop()

	outputFile, bufWriter, err := e.createBufferedOutputFile(outputPath, permissions)
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

	cmd := exec.CommandContext(ctx, dumpBin.Path, mysqldumpArgs...)

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

		// Extract exit code dari error
		exitCode := 1 // default exit code untuk generic error
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}

		if e.isFatalDumpError(runErr, stderrOutput, exitCode) {
			result := &types_backup.BackupWriteResult{
				StderrOutput: stderrOutput,
				FileSize:     0,
			}
			excerpt := summarizeStderr(stderrOutput, 20, 2000)
			if excerpt != "" {
				return result, fmt.Errorf("%s gagal: %w (stderr: %s)", dumpBin.Name, runErr, excerpt)
			}
			return result, fmt.Errorf("%s gagal: %w", dumpBin.Name, runErr)
		}
		excerpt := summarizeStderr(stderrOutput, 12, 1200)
		if excerpt != "" {
			e.Log.Warnf("%s exit with non-fatal error, treated as warning. stderr (excerpt):\n%s", dumpBin.Name, excerpt)
		} else {
			e.Log.Warnf("%s exit with non-fatal error, treated as warning", dumpBin.Name)
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
