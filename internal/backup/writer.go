// File : internal/backup/writer.go
// Deskripsi : Writer dan mysqldump execution logic
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package backup

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/backuphelper"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"
)

const bufferSize = 256 * 1024 // 256KB buffer untuk buffered I/O

// executeMysqldumpWithPipe menjalankan mysqldump dengan pipe untuk kompresi dan enkripsi
func (s *Service) executeMysqldumpWithPipe(ctx context.Context, mysqldumpArgs []string, outputPath string, compressionRequired bool, compressionType string) (*types_backup.BackupWriteResult, error) {
	// Resolve encryption key SEBELUM spinner dimulai
	var encryptionKey string
	if s.BackupDBOptions.Encryption.Enabled {
		resolvedKey, source, err := helper.ResolveEncryptionKey(s.BackupDBOptions.Encryption.Key, consts.ENV_BACKUP_ENCRYPTION_KEY)
		if err != nil {
			return nil, fmt.Errorf("gagal mendapatkan kunci enkripsi: %w", err)
		}
		encryptionKey = resolvedKey
		s.Log.Debugf("Kunci enkripsi didapat dari: %s", source)
	}

	// Start spinner dengan elapsed time
	spin := ui.NewSpinnerWithElapsed("Memproses backup database")
	spin.Start()
	defer spin.Stop()

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat file output: %w", err)
	}
	defer outputFile.Close()

	// Setup buffered writer
	bufWriter := bufio.NewWriterSize(outputFile, bufferSize)
	defer func() {
		if flushErr := bufWriter.Flush(); flushErr != nil {
			s.Log.Errorf("Error flushing buffer: %v", flushErr)
		}
	}()

	// Setup writer chain: mysqldump → Compression → Encryption → Buffer → File
	writer, closers, err := s.createWriterPipeline(bufWriter, compressionRequired, compressionType, encryptionKey)
	if err != nil {
		return nil, err
	}

	// Cleanup writers
	defer func() {
		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				s.Log.Errorf("Error closing writer: %v", err)
			}
		}
	}()

	// Execute mysqldump command
	cmd := exec.CommandContext(ctx, "mysqldump", mysqldumpArgs...)
	cmd.Stdout = writer

	// Capture stderr
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	// Run command
	if err := cmd.Run(); err != nil {
		stderrOutput := stderrBuf.String()

		// Log ke error log file
		logFile := s.ErrorLog.LogWithOutput(map[string]interface{}{
			"type": "mysqldump_backup",
			"file": outputPath,
		}, stderrOutput, err)
		_ = logFile

		// Check fatal error
		if s.isFatalMysqldumpError(err, stderrOutput) {
			return nil, fmt.Errorf("mysqldump gagal: %w", err)
		}
		s.Log.Warn("mysqldump exit with non-fatal error, treated as warning")
	}

	// Get file info
	fileInfo, statErr := os.Stat(outputPath)
	var fileSize int64
	if statErr == nil {
		fileSize = fileInfo.Size()
	}

	// Create result
	result := &types_backup.BackupWriteResult{
		StderrOutput: stderrBuf.String(),
		FileSize:     fileSize,
	}

	return result, nil
}

// createWriterPipeline membuat writer pipeline untuk compression dan encryption
func (s *Service) createWriterPipeline(baseWriter io.Writer, compressionRequired bool, compressionType string, encryptionKey string) (io.Writer, []io.Closer, error) {
	var writer io.Writer = baseWriter
	var closers []io.Closer

	// Layer 1: Encryption (paling dekat dengan file)
	if s.BackupDBOptions.Encryption.Enabled {
		encryptingWriter, err := encrypt.NewEncryptingWriter(writer, []byte(encryptionKey))
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
			Level: compress.CompressionLevel(s.BackupDBOptions.Compression.Level),
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

// isFatalMysqldumpError menentukan apakah error dari mysqldump adalah fatal
func (s *Service) isFatalMysqldumpError(err error, stderrOutput string) bool {
	if err == nil {
		return false
	}

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
