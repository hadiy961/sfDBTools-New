package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/encrypt"
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

	// Start spinner
	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	spin.Suffix = " Memproses backup database..."
	spin.Start()
	defer spin.Stop()

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
			resolvedKey, source, err := helper.ResolveEncryptionKey(s.BackupDBOptions.Encryption.Key, consts.ENV_BACKUP_ENCRYPTION_KEY)
			if err != nil {
				return nil, fmt.Errorf("gagal mendapatkan kunci enkripsi: %w", err)
			}
			encryptionKey = resolvedKey
			s.Log.Infof("Kunci enkripsi diperoleh dari: %s", source)
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

		// Log error untuk debugging
		s.Log.Errorf("mysqldump exit with error: %v", err)
		s.Log.Errorf("mysqldump stderr: %s", stderrOutput)

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
	if err == nil {
		return false
	}

	// Jika stderr kosong tapi ada error, anggap fatal
	if stderrOutput == "" {
		s.Log.Debug("mysqldump error with empty stderr, treating as fatal")
		return true
	}

	stderrLower := strings.ToLower(stderrOutput)

	// Cek pattern error yang FATAL terlebih dahulu
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

	for _, pattern := range fatalPatterns {
		if strings.Contains(stderrLower, pattern) {
			// s.Log.Debugf("mysqldump stderr contains fatal pattern: %s", pattern)
			return true
		}
	}

	// Cek exit code jika tersedia
	// Exit code 2 BISA jadi warning atau error, tergantung stderr
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode := exitErr.ExitCode()
		// mysqldump exit code 0 = success, 2 = bisa warning atau error
		if exitCode == 2 {
			// Jika exit code 2 dan tidak ada fatal pattern di atas, anggap warning
			s.Log.Debugf("mysqldump exit code 2, no fatal pattern found, treating as non-fatal")
		}
	}

	// Cek pattern error yang non-fatal (warnings yang tidak menghentikan dump)
	nonFatalPatterns := []string{
		"couldn't read keys from table",
		"references invalid table(s) or column(s) or function(s)",
		"definer/invoker of view lack rights",
		"warning:",
	}

	for _, pattern := range nonFatalPatterns {
		if strings.Contains(stderrLower, pattern) {
			s.Log.Debugf("mysqldump stderr contains non-fatal pattern: %s", pattern)
			return false
		}
	}

	// Default: jika ada error dan tidak match pattern apapun, anggap fatal
	s.Log.Debug("mysqldump error doesn't match any pattern, treating as fatal")
	return true
}

// maskPasswordInArgs mem-mask password di mysqldump arguments untuk logging
func (s *Service) maskPasswordInArgs(args []string) []string {
	masked := make([]string, len(args))
	copy(masked, args)

	for i := 0; i < len(masked); i++ {
		arg := masked[i]

		// Cek format -pPASSWORD atau --password=PASSWORD
		if strings.HasPrefix(arg, "-p") && len(arg) > 2 {
			// Format: -pPASSWORD
			masked[i] = "-p********"
		} else if strings.HasPrefix(arg, "--password=") {
			// Format: --password=PASSWORD
			masked[i] = "--password=********"
		} else if arg == "-p" || arg == "--password" {
			// Format: -p PASSWORD atau --password PASSWORD (password di arg berikutnya)
			if i+1 < len(masked) {
				masked[i+1] = "********"
			}
		}
	}

	return masked
}
