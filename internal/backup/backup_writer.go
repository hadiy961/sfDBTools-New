package backup

import (
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
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/helper"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

// executeMysqldumpWithPipe menjalankan mysqldump dengan pipe untuk kompresi dan enkripsi.
// Mengembalikan BackupWriteResult yang berisi stderr output dan checksums
func (s *Service) executeMysqldumpWithPipe(ctx context.Context, mysqldumpArgs []string, outputPath string, compressionRequired bool, compressionType string) (*types.BackupWriteResult, error) {
	// Mask password untuk logging
	// maskedArgs := s.maskPasswordInArgs(mysqldumpArgs)

	// Start spinner dengan elapsed time
	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	spin.Suffix = " Memproses backup database..."
	spin.Start()

	// Goroutine untuk update spinner dengan elapsed time
	startTime := time.Now()
	done := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				elapsed := time.Since(startTime)
				spin.Suffix = fmt.Sprintf(" Memproses backup database... (%s)", global.FormatDuration(elapsed))
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	defer func() {
		done <- true // Stop elapsed time updater
		spin.Stop()
	}()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat file output: %w", err)
	}
	defer outputFile.Close()

	// Setup writer chain: mysqldump → Compression → Encryption → File
	// HashWriter akan di-setup sebagai parallel writer untuk hash dari raw data
	var writer io.Writer = outputFile
	var hashWriter *MultiHashWriter
	var closers []io.Closer

	// Layer 1: Encryption (paling dekat dengan file)
	if s.BackupDBOptions.Encryption.Enabled {
		encryptionKey := s.BackupDBOptions.Encryption.Key
		if encryptionKey == "" {
			resolvedKey, _, err := helper.ResolveEncryptionKey(s.BackupDBOptions.Encryption.Key, consts.ENV_BACKUP_ENCRYPTION_KEY)
			if err != nil {
				return nil, fmt.Errorf("gagal mendapatkan kunci enkripsi: %w", err)
			}
			encryptionKey = resolvedKey
		}

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

	// Layer 3: Setup HashWriter sebagai TeeWriter
	// Data dari mysqldump akan ditulis ke DUA tempat secara parallel:
	// 1. Ke writer chain (compression → encryption → file)
	// 2. Ke hashWriter untuk kalkulasi checksum
	// Dengan cara ini, hash dihitung dari data SQL mentah SEBELUM compression/encryption
	if s.Config.Backup.Verification.CompareChecksums {
		hashWriter = NewMultiHashWriter(io.Discard) // Hash saja, tidak tulis ke file
		// Gunakan TeeWriter untuk split data ke 2 destination
		writer = io.MultiWriter(writer, hashWriter)
	}

	// cmd.Stdout akan write ke writer:
	// mysqldump → MultiWriter → [compressingWriter → encryptingWriter → file] + [hashWriter]
	// Data flow: mysqldump → hashWriter → compressingWriter → encryptingWriter → file

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

	// Buat result dengan checksums jika enabled
	result := &types.BackupWriteResult{
		StderrOutput: stderrBuf.String(),
	}

	if hashWriter != nil {
		result.SHA256Hash = hashWriter.GetSHA256()
		result.MD5Hash = hashWriter.GetMD5()
		result.BytesWritten = hashWriter.GetBytesWritten()
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
