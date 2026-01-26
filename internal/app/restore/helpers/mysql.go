// File : internal/restore/helpers/mysql.go
// Deskripsi : Helper functions untuk MySQL restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 17 Desember 2025
// Last Modified : 26 Januari 2026
package helpers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	backupfile "sfdbtools/internal/app/backup/helpers/file"
	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/compress"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/progress"
	"strings"
)

func resolveMariaDBOrMySQLClient() (binPath string, binName string, err error) {
	// Default: mariadb client (mysql CLI compatible)
	if p, e := exec.LookPath("mariadb"); e == nil {
		return p, "mariadb", nil
	}
	// Fallback: mysql client
	if p, e := exec.LookPath("mysql"); e == nil {
		return p, "mysql", nil
	}
	return "", "", fmt.Errorf("binary client database tidak ditemukan: butuh 'mariadb' atau 'mysql' di PATH")
}

func isSSLMismatchServerNotSupport(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "tls/ssl error") && strings.Contains(msg, "server does not support")
}

func hasSkipSSLArg(args []string) bool {
	for _, a := range args {
		if strings.TrimSpace(strings.ToLower(a)) == "--skip-ssl" {
			return true
		}
	}
	return false
}

// BuildMySQLArgs membuat argument list untuk mysql command
func BuildMySQLArgs(profile *domain.ProfileInfo, database string, extraArgs ...string) []string {
	eff := profileconn.EffectiveDBInfo(profile)
	args := []string{
		fmt.Sprintf("--host=%s", eff.Host),
		fmt.Sprintf("--port=%d", eff.Port),
		fmt.Sprintf("--user=%s", profile.DBInfo.User),
		fmt.Sprintf("--password=%s", profile.DBInfo.Password),
	}

	// Tambahkan extra args jika ada
	args = append(args, extraArgs...)

	// Tambahkan database jika specified
	if database != "" {
		args = append(args, database)
	}

	return args
}

// ExecuteMySQLCommand menjalankan mysql command dengan stdin reader
func ExecuteMySQLCommand(ctx context.Context, args []string, stdin io.Reader) error {
	binPath, binName, err := resolveMariaDBOrMySQLClient()
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, binPath, args...)
	cmd.Stdin = stdin

	var stderr strings.Builder
	cmd.Stderr = &stderr
	cmd.Stdout = io.Discard

	if err := cmd.Run(); err != nil {
		stderrMsg := stderr.String()
		if stderrMsg != "" {
			return fmt.Errorf("%s command error: %w (stderr: %s)", binName, err, stderrMsg)
		}
		return fmt.Errorf("%s command error: %w", binName, err)
	}

	return nil
}

// RestoreFromFile melakukan restore database dari file backup
func RestoreFromFile(ctx context.Context, filePath string, targetDB string, profile *domain.ProfileInfo, encryptionKey string) error {
	spin := progress.NewSpinnerWithElapsed(fmt.Sprintf("Restore database %s dari %s", targetDB, filepath.Base(filePath)))
	spin.Start()
	defer spin.Stop()

	// Build mysql args dengan force flag
	args := BuildMySQLArgs(profile, targetDB, "-f")

	// Helper closure supaya retry bisa reopen file (stdin streaming tidak bisa diulang).
	execRestore := func(a []string) error {
		reader, closers, err := OpenAndPrepareReader(filePath, encryptionKey)
		if err != nil {
			return err
		}
		defer CloseReaders(closers)
		return ExecuteMySQLCommand(ctx, a, reader)
	}

	// Execute mysql restore
	if err := execRestore(args); err != nil {
		// Fallback: beberapa environment punya default SSL=ON/REQUIRED di client config.
		// Jika target server tidak support SSL, retry sekali dengan SSL dimatikan.
		if isSSLMismatchServerNotSupport(err) && !hasSkipSSLArg(args) {
			retryArgs := BuildMySQLArgs(profile, targetDB, "--skip-ssl", "-f")
			if err2 := execRestore(retryArgs); err2 == nil {
				return nil
			} else {
				return fmt.Errorf("gagal menjalankan mysql restore (retry --skip-ssl): %w", err2)
			}
		}
		return fmt.Errorf("gagal menjalankan mysql restore: %w", err)
	}

	return nil
}

// OpenAndPrepareReader membuka file dan menyiapkan reader dengan decrypt/decompress
// Returns: reader, list of closers, error
func OpenAndPrepareReader(filePath string, encryptionKey string) (io.Reader, []io.Closer, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal membuka file: %w", err)
	}

	// Buffer file reads to improve large sequential throughput
	reader := io.Reader(bufio.NewReaderSize(file, 4*1024*1024))
	closers := []io.Closer{file}

	// Decrypt if encrypted
	isEncrypted := backupfile.IsEncryptedFile(filePath)
	if isEncrypted {
		decReader, err := crypto.NewStreamDecryptor(reader, encryptionKey)
		if err != nil {
			CloseReaders(closers)
			return nil, nil, fmt.Errorf("gagal membuat decrypting reader: %w", err)
		}
		reader = decReader
		closers = append(closers, io.NopCloser(decReader))
	}

	// Decompress if compressed
	compressionType := compress.DetectCompressionTypeFromFile(filePath)
	if compressionType != compress.CompressionType(consts.CompressionTypeNone) {
		decompReader, err := compress.NewDecompressingReader(reader, compressionType)
		if err != nil {
			CloseReaders(closers)
			return nil, nil, fmt.Errorf("gagal membuat decompressing reader: %w", err)
		}
		reader = decompReader
		closers = append(closers, decompReader)
	}

	return reader, closers, nil
}

// CloseReaders menutup semua readers dengan urutan terbalik
func CloseReaders(closers []io.Closer) {
	for i := len(closers) - 1; i >= 0; i-- {
		if closer := closers[i]; closer != nil {
			_ = closer.Close()
		}
	}
}

// RestoreUserGrants melakukan restore user grants dari file
func RestoreUserGrants(ctx context.Context, grantsFile string, profile *domain.ProfileInfo) error {
	if grantsFile == "" {
		return nil
	}

	spin := progress.NewSpinnerWithElapsed(fmt.Sprintf("Restore user grants dari %s", filepath.Base(grantsFile)))
	spin.Start()
	defer spin.Stop()

	grantsSQL, err := os.ReadFile(grantsFile)
	if err != nil {
		return fmt.Errorf("gagal membaca file grants: %w", err)
	}

	// Build mysql args tanpa database target
	args := BuildMySQLArgs(profile, "")

	// Execute mysql restore
	if err := ExecuteMySQLCommand(ctx, args, strings.NewReader(string(grantsSQL))); err != nil {
		if isSSLMismatchServerNotSupport(err) && !hasSkipSSLArg(args) {
			retryArgs := BuildMySQLArgs(profile, "", "--skip-ssl")
			if err2 := ExecuteMySQLCommand(ctx, retryArgs, strings.NewReader(string(grantsSQL))); err2 == nil {
				return nil
			} else {
				return fmt.Errorf("gagal restore user grants (retry --skip-ssl): %w", err2)
			}
		}
		return fmt.Errorf("gagal restore user grants: %w", err)
	}

	return nil
}
