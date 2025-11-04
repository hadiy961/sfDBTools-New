package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/helper"
	"strings"
)

// executeMysqldumpWithPipe menjalankan mysqldump dengan pipe untuk kompresi dan enkripsi.
// Mengembalikan error untuk fatal errors dan stderr output untuk warnings/non-fatal errors
func (s *Service) executeMysqldumpWithPipe(ctx context.Context, mysqldumpArgs []string, outputPath string, compressionRequired bool, compressionType string) (string, error) {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file output: %w", err)
	}
	defer outputFile.Close()

	var writer io.Writer = outputFile
	var closers []io.Closer

	// Urutan layer: mysqldump -> Compression -> Encryption -> File
	if s.BackupDBOptions.Encryption.Enabled {
		encryptionKey := s.BackupDBOptions.Encryption.Key
		if encryptionKey == "" {
			resolvedKey, source, err := helper.ResolveEncryptionKey(s.BackupDBOptions.Encryption.Key, consts.ENV_BACKUP_ENCRYPTION_KEY)
			if err != nil {
				return "", fmt.Errorf("gagal mendapatkan kunci enkripsi: %w", err)
			}
			encryptionKey = resolvedKey
			s.Log.Infof("Kunci enkripsi diperoleh dari: %s", source)
		}

		encryptingWriter, err := encrypt.NewEncryptingWriter(writer, []byte(encryptionKey))
		if err != nil {
			return "", fmt.Errorf("gagal membuat encrypting writer: %w", err)
		}
		closers = append(closers, encryptingWriter)
		writer = encryptingWriter
	}

	if compressionRequired {
		compressionConfig := compress.CompressionConfig{
			Type:  compress.CompressionType(compressionType),
			Level: compress.CompressionLevel(s.BackupDBOptions.Compression.Level),
		}
		compressingWriter, err := compress.NewCompressingWriter(writer, compressionConfig)
		if err != nil {
			return "", fmt.Errorf("gagal membuat compressing writer: %w", err)
		}
		closers = append(closers, compressingWriter)
		writer = compressingWriter
	}

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
		// Cek apakah ini error fatal atau hanya warning
		if s.isFatalMysqldumpError(err, stderrOutput) {
			return stderrOutput, fmt.Errorf("mysqldump gagal: %w", err)
		}
		// Jika bukan fatal error, kembalikan stderr sebagai warning
		return stderrOutput, nil
	}

	return stderrBuf.String(), nil
}

// isFatalMysqldumpError menentukan apakah error dari mysqldump adalah fatal atau hanya warning
// Fatal errors: koneksi gagal, permission denied, database tidak ada, dll
// Non-fatal: view errors, trigger errors (data masih bisa di-backup)
func (s *Service) isFatalMysqldumpError(err error, stderrOutput string) bool {
	if err == nil {
		return false
	}

	// Cek exit code jika tersedia
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode := exitErr.ExitCode()
		// mysqldump exit code 0 = success, 2 = warning (non-fatal)
		if exitCode == 2 {
			return false
		}
	}

	// Cek pattern error yang non-fatal (warnings yang tidak menghentikan dump)
	nonFatalPatterns := []string{
		"Couldn't read keys from table",
		"references invalid table(s) or column(s) or function(s)",
		"definer/invoker of view lack rights",
		"Warning:",
	}

	stderrLower := strings.ToLower(stderrOutput)
	for _, pattern := range nonFatalPatterns {
		if strings.Contains(stderrLower, strings.ToLower(pattern)) {
			// Jika hanya ada warning patterns dan bukan error fatal lainnya
			// Cek juga apakah ada error fatal
			if !strings.Contains(stderrLower, "access denied") &&
				!strings.Contains(stderrLower, "unknown database") &&
				!strings.Contains(stderrLower, "can't connect") &&
				!strings.Contains(stderrLower, "connection refused") {
				return false
			}
		}
	}

	// Default: anggap fatal
	return true
}
