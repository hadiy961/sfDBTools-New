package backup

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/backuphelper"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"
)

const (
	// bufferSize untuk buffered I/O (256KB)
	bufferSize = 256 * 1024
)

// executeMysqldumpWithPipe menjalankan mysqldump dengan pipe untuk kompresi dan enkripsi.
// Mengembalikan BackupWriteResult yang berisi stderr output dan checksums
func (s *Service) executeMysqldumpWithPipe(ctx context.Context, mysqldumpArgs []string, outputPath string, compressionRequired bool, compressionType string) (*types.BackupWriteResult, error) {
	// Mask password untuk logging
	// maskedArgs := s.maskPasswordInArgs(mysqldumpArgs)

	// PENTING: Resolve encryption key SEBELUM spinner dimulai
	// Jika tidak ada env var, ini akan meminta input dari user via prompt
	var encryptionKey string
	if s.BackupDBOptions.Encryption.Enabled {
		resolvedKey, source, err := helper.ResolveEncryptionKey(s.BackupDBOptions.Encryption.Key, consts.ENV_BACKUP_ENCRYPTION_KEY)
		if err != nil {
			return nil, fmt.Errorf("gagal mendapatkan kunci enkripsi: %w", err)
		}
		encryptionKey = resolvedKey
		s.Log.Debugf("Kunci enkripsi didapat dari: %s", source)
	}

	// Start spinner dengan elapsed time SETELAH encryption key resolved
	spin := ui.NewSpinnerWithElapsed("Memproses backup database")
	spin.Start()
	defer spin.Stop()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat file output: %w", err)
	}
	defer outputFile.Close()

	// Tambahkan buffered writer untuk mengurangi syscall overhead
	bufWriter := bufio.NewWriterSize(outputFile, bufferSize)
	defer func() {
		if flushErr := bufWriter.Flush(); flushErr != nil {
			s.Log.Errorf("Error flushing buffer: %v", flushErr)
		}
	}()

	// Setup writer chain: mysqldump → Compression → Encryption → Buffer → File
	var writer io.Writer = bufWriter
	var closers []io.Closer

	// Layer 1: Encryption (paling dekat dengan file)
	if s.BackupDBOptions.Encryption.Enabled {
		encryptingWriter, err := encrypt.NewEncryptingWriter(writer, []byte(encryptionKey))
		if err != nil {
			return nil, fmt.Errorf("gagal membuat encrypting writer: %w", err)
		}
		closers = append(closers, encryptingWriter)
		writer = encryptingWriter
	}

	// Layer 2: Compression
	if compressionRequired {
		compressionConfig := compress.CompressionConfig{
			Type:  compress.CompressionType(compressionType),
			Level: compress.CompressionLevel(s.BackupDBOptions.Compression.Level),
		}
		compressingWriter, err := compress.NewCompressingWriter(writer, compressionConfig)
		if err != nil {
			return nil, fmt.Errorf("gagal membuat compressing writer: %w", err)
		}
		closers = append(closers, compressingWriter)
		writer = compressingWriter
	}

	// cmd.Stdout akan write ke writer:
	// mysqldump → compressingWriter → encryptingWriter → file

	defer func() {
		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				s.Log.Errorf("Error closing writer: %v", err)
			}
		}
	}()

	cmd := exec.CommandContext(ctx, "mysqldump", mysqldumpArgs...)
	cmd.Stdout = writer

	// Capture stderr untuk menangkap warnings dan errors
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	// logArgs := s.sanitizeArgsForLogging(mysqldumpArgs)
	// s.Logger.Infof("Command: mysqldump %s", strings.Join(logArgs, " "))

	if err := cmd.Run(); err != nil {
		stderrOutput := stderrBuf.String()

		// Log ke error log file terpisah
		logFile := s.ErrorLog.LogWithOutput(map[string]interface{}{
			"type": "mysqldump_backup",
			"file": outputPath,
		}, stderrOutput, err)
		_ = logFile

		// Cek apakah ini error fatal atau hanya warning
		if s.isFatalMysqldumpError(err, stderrOutput) {
			return nil, fmt.Errorf("mysqldump gagal: %w", err)
		}
		// Jika bukan fatal error, kembalikan stderr sebagai warning
		s.Log.Warn("mysqldump exit with non-fatal error, treated as warning")
	}

	// Buat result
	result := &types.BackupWriteResult{
		StderrOutput: stderrBuf.String(),
	}

	return result, nil
}

// isFatalMysqldumpError menentukan apakah error dari mysqldump adalah fatal atau hanya warning
// Fatal errors: koneksi gagal, permission denied, database tidak ada, dll
// Non-fatal: view errors, trigger errors (data masih bisa di-backup)
func (s *Service) isFatalMysqldumpError(err error, stderrOutput string) bool {
	// Delegate to package helper for centralized heuristics
	if err == nil {
		return false
	}

	// If stderr empty treat as fatal (keep the same logging behaviour)
	if stderrOutput == "" {
		s.Log.Debug("mysqldump error with empty stderr, treating as fatal")
		return true
	}

	fatal := backuphelper.IsFatalMysqldumpError(err, stderrOutput)

	if !fatal {
		s.Log.Debugf("mysqldump treated as non-fatal by helper: %s", stderrOutput)
	}

	return fatal
}

// maskPasswordInArgs mem-mask password di mysqldump arguments untuk logging
func (s *Service) maskPasswordInArgs(args []string) []string {
	return backuphelper.MaskPasswordInArgs(args)
}
