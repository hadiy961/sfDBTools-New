// File : internal/crypto/file/encrypt.go
// Deskripsi : File encryption convenience functions
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

// Threshold untuk memilih metode enkripsi (1MB)
// File < 1MB: gunakan simple format (compatible dengan profile)
// File ≥ 1MB: gunakan streaming format (efficient untuk backup)
const encryptionThresholdBytes = 1024 * 1024

// Encrypt encrypts a file using auto-selected encryption method.
//
// Automatically selects encryption method based on file size:
//   - Files < 1MB: Simple format (compatible with profile files)
//   - Files ≥ 1MB: Streaming format (efficient for large backups)
//
// Both formats use AES-256-GCM and are OpenSSL compatible.
//
// Parameters:
//   - inputPath: path to plaintext file
//   - outputPath: path to encrypted output file
//   - passphrase: encryption password/key
//
// Returns error if encryption fails.
//
// Example:
//
//	err := file.Encrypt("data.txt", "data.txt.enc", []byte("my-password"))
func Encrypt(inputPath, outputPath string, passphrase []byte) error {
	// Get file size to determine encryption method
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to stat input file: %w", err)
	}

	fileSize := fileInfo.Size()

	// Small files: use simple format (compatible dengan profile)
	if fileSize < encryptionThresholdBytes {
		return encryptSimple(inputPath, outputPath, passphrase)
	}

	// Large files: use streaming format
	return encryptStreaming(inputPath, outputPath, passphrase)
}

// encryptSimple encrypts small files using in-memory encryption.
// Compatible with profile files and DecryptData().
func encryptSimple(inputPath, outputPath string, passphrase []byte) error {
	// Read entire file to memory
	plaintext, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}
	// Zero plaintext memory after encryption
	defer func() {
		for i := range plaintext {
			plaintext[i] = 0
		}
	}()

	// Encrypt using simple format
	ciphertext, err := core.EncryptAES(plaintext, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Write to output file
	if err := os.WriteFile(outputPath, ciphertext, core.SecureFilePermission); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// encryptStreaming encrypts large files using streaming encryption.
// Efficient for large backup files.
func encryptStreaming(inputPath, outputPath string, passphrase []byte) error {
	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Create output file with secure permissions atomically
	outputFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, core.SecureFilePermission)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Create encrypting writer
	encWriter, err := stream.NewWriter(outputFile, passphrase)
	if err != nil {
		return fmt.Errorf("failed to create encrypting writer: %w", err)
	}
	defer encWriter.Close()

	// Show progress for large files
	sp := progress.NewSpinnerWithElapsed("Mengenkripsi file (streaming)")
	sp.Start()
	defer sp.Stop()

	// Stream copy with encryption
	if _, err := io.Copy(encWriter, inputFile); err != nil {
		return fmt.Errorf("failed to encrypt file: %w", err)
	}

	return nil
}
