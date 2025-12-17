// File : internal/restore/helpers/mysql.go
// Deskripsi : Helper functions untuk MySQL restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package helpers

import (
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

// RestoreFromFile melakukan restore database dari file backup
func RestoreFromFile(ctx context.Context, filePath string, targetDB string, profile *types.ProfileInfo, encryptionKey string) error {
	spin := ui.NewSpinnerWithElapsed(fmt.Sprintf("Restore database %s dari %s", targetDB, filepath.Base(filePath)))
	spin.Start()
	defer spin.Stop()

	backupFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("gagal membuka file backup: %w", err)
	}
	defer backupFile.Close()

	reader := io.Reader(backupFile)
	var closers []io.Closer

	// Decrypt if encrypted
	isEncrypted := strings.HasSuffix(strings.ToLower(filePath), ".enc")
	if isEncrypted {
		decReader, err := encrypt.NewDecryptingReader(reader, encryptionKey)
		if err != nil {
			return fmt.Errorf("gagal membuat decrypting reader: %w", err)
		}
		reader = decReader
		closers = append(closers, io.NopCloser(decReader))
	}

	// Decompress if compressed
	compressionType := compress.DetectCompressionTypeFromFile(filePath)
	if compressionType != compress.CompressionNone {
		decompReader, err := compress.NewDecompressingReader(reader, compressionType)
		if err != nil {
			return fmt.Errorf("gagal membuat decompressing reader: %w", err)
		}
		reader = decompReader
		closers = append(closers, decompReader)
	}

	defer func() {
		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				// Log error if logger available
			}
		}
	}()

	// Execute mysql restore
	args := []string{
		fmt.Sprintf("--host=%s", profile.DBInfo.Host),
		fmt.Sprintf("--port=%d", profile.DBInfo.Port),
		fmt.Sprintf("--user=%s", profile.DBInfo.User),
		fmt.Sprintf("--password=%s", profile.DBInfo.Password),
		"-f", // Force continue on error
		targetDB,
	}

	cmd := exec.CommandContext(ctx, "mysql", args...)
	cmd.Stdin = reader

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gagal menjalankan mysql restore: %w (stderr: %s)", err, stderr.String())
	}

	return nil
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

	args := []string{
		fmt.Sprintf("--host=%s", profile.DBInfo.Host),
		fmt.Sprintf("--port=%d", profile.DBInfo.Port),
		fmt.Sprintf("--user=%s", profile.DBInfo.User),
		fmt.Sprintf("--password=%s", profile.DBInfo.Password),
	}

	cmd := exec.CommandContext(ctx, "mysql", args...)
	cmd.Stdin = strings.NewReader(string(grantsSQL))

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := stderr.String()
		if stderrMsg != "" {
			return fmt.Errorf("gagal restore user grants: %w (stderr: %s)", err, stderrMsg)
		}
		return fmt.Errorf("gagal restore user grants: %w", err)
	}

	return nil
}
