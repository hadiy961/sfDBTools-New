// File : internal/restore/helpers/mysql.go
// Deskripsi : Helper functions untuk MySQL restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package helpers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/ui"
	"strings"
)

// BuildMySQLArgs membuat argument list untuk mysql command
func BuildMySQLArgs(profile *types.ProfileInfo, database string, extraArgs ...string) []string {
	args := []string{
		fmt.Sprintf("--host=%s", profile.DBInfo.Host),
		fmt.Sprintf("--port=%d", profile.DBInfo.Port),
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
	cmd := exec.CommandContext(ctx, "mysql", args...)
	cmd.Stdin = stdin

	var stderr strings.Builder
	cmd.Stderr = &stderr
	cmd.Stdout = io.Discard

	if err := cmd.Run(); err != nil {
		stderrMsg := stderr.String()
		if stderrMsg != "" {
			return fmt.Errorf("mysql command error: %w (stderr: %s)", err, stderrMsg)
		}
		return fmt.Errorf("mysql command error: %w", err)
	}

	return nil
}

// RestoreFromFile melakukan restore database dari file backup
func RestoreFromFile(ctx context.Context, filePath string, targetDB string, profile *types.ProfileInfo, encryptionKey string) error {
	spin := ui.NewSpinnerWithElapsed(fmt.Sprintf("Restore database %s dari %s", targetDB, filepath.Base(filePath)))
	spin.Start()
	defer spin.Stop()

	// Open, decrypt, decompress file
	reader, closers, err := OpenAndPrepareReader(filePath, encryptionKey)
	if err != nil {
		return err
	}
	defer CloseReaders(closers)

	// Build mysql args dengan force flag
	args := BuildMySQLArgs(profile, targetDB, "-f")

	// Execute mysql restore
	if err := ExecuteMySQLCommand(ctx, args, reader); err != nil {
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
	isEncrypted := strings.HasSuffix(strings.ToLower(filePath), ".enc")
	if isEncrypted {
		decReader, err := encrypt.NewDecryptingReader(reader, encryptionKey)
		if err != nil {
			CloseReaders(closers)
			return nil, nil, fmt.Errorf("gagal membuat decrypting reader: %w", err)
		}
		reader = decReader
		closers = append(closers, io.NopCloser(decReader))
	}

	// Decompress if compressed
	compressionType := compress.DetectCompressionTypeFromFile(filePath)
	if compressionType != compress.CompressionNone {
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
func RestoreUserGrants(ctx context.Context, grantsFile string, profile *types.ProfileInfo) error {
	if grantsFile == "" {
		return nil
	}

	spin := ui.NewSpinnerWithElapsed(fmt.Sprintf("Restore user grants dari %s", filepath.Base(grantsFile)))
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
		return fmt.Errorf("gagal restore user grants: %w", err)
	}

	return nil
}
