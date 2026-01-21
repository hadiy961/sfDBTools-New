// File : internal/crypto/file/decrypt.go
// Deskripsi : File decryption convenience functions
// Author : Hadiyatna Muflihun
// Tanggal : 8 Januari 2026
// Last Modified : 21 Januari 2026
package file

import (
	"fmt"
	"io"
	"os"

	"sfdbtools/internal/crypto/core"
	"sfdbtools/internal/crypto/stream"
	"sfdbtools/internal/ui/progress"
)

// Decrypt decrypts a file using auto-detected decryption method.
//
// Automatically detects encryption format:
//   - Tries simple format first (profile files, small files)
//   - Falls back to streaming format (large backup files)
//
// This makes the function compatible with both:
//   - Profile files (.cnf.enc) encrypted with EncryptData
//   - Backup files encrypted with streaming encryption
//
// Parameters:
//   - inputPath: path to encrypted file
//   - outputPath: path to decrypted output file
//   - passphrase: decryption password/key
//
// Returns error if decryption fails with both methods.
//
// Example:
//
//	err := file.Decrypt("data.txt.enc", "data.txt", []byte("my-password"))
func Decrypt(inputPath, outputPath string, passphrase []byte) error {
	// Try simple format first (for profile files and small files)
	err := decryptSimple(inputPath, outputPath, passphrase)
	if err == nil {
		return nil
	}

	// Fallback to streaming format (for large backup files)
	err = decryptStreaming(inputPath, outputPath, passphrase)
	if err == nil {
		return nil
	}

	// Both methods failed - provide actionable error message
	return fmt.Errorf("failed to decrypt file: %w\n\nPossible causes:\n  - Incorrect password/key\n  - File is corrupted or not encrypted\n  - File was encrypted with different tool\n\nHint: Verify password and try: sfdbtools crypto decrypt-file --help", err)
}

// decryptSimple decrypts files using in-memory decryption.
// Compatible with profile files and EncryptData().
func decryptSimple(inputPath, outputPath string, passphrase []byte) error {
	// Read entire encrypted file to memory
	ciphertext, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}

	// Decrypt using simple format
	plaintext, err := core.DecryptAES(ciphertext, passphrase)
	if err != nil {
		return fmt.Errorf("simple format decryption failed: %w", err)
	}
	// Zero plaintext memory after writing
	defer func() {
		for i := range plaintext {
			plaintext[i] = 0
		}
	}()

	// Write to temp file first untuk atomic operation
	tmpFile, err := os.CreateTemp("", "sfdbtools-decrypt-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	shouldCleanup := true
	defer func() {
		if shouldCleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if err := tmpFile.Chmod(core.SecureFilePermission); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to set temp file permissions: %w", err)
	}

	if _, err := tmpFile.Write(plaintext); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic rename dari temp ke output
	if err := os.Rename(tmpPath, outputPath); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	shouldCleanup = false // Setelah rename sukses, file sudah dipindahkan

	return nil
}

// decryptStreaming decrypts files using streaming decryption.
// Efficient for large backup files.
func decryptStreaming(inputPath, outputPath string, passphrase []byte) error {
	// Open encrypted file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open encrypted file: %w", err)
	}
	defer inputFile.Close()

	// Create decrypting reader
	decReader, err := stream.NewReader(inputFile, string(passphrase))
	if err != nil {
		return fmt.Errorf("failed to create decrypting reader: %w", err)
	}

	// Create output file with secure permissions atomically
	outputFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, core.SecureFilePermission)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Show progress for large files
	sp := progress.NewSpinnerWithElapsed("Mendekripsi file (streaming)")
	sp.Start()
	defer sp.Stop()

	// Stream copy with decryption
	if _, err := io.Copy(outputFile, decReader); err != nil {
		return fmt.Errorf("streaming format decryption failed: %w", err)
	}

	return nil
}
