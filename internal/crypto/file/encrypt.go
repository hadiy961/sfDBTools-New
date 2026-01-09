// File : internal/crypto/file/encrypt.go
// Deskripsi : File encryption convenience functions
// Author : Hadiyatna Muflihun
// Tanggal : 8 Januari 2026
// Last Modified : 8 Januari 2026
package file

import (
	"fmt"
	"io"
	"os"

	"sfdbtools/internal/crypto/stream"
)

// Encrypt encrypts a file using streaming encryption.
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
	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Create output file
	outputFile, err := os.Create(outputPath)
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

	// Stream copy with encryption
	if _, err := io.Copy(encWriter, inputFile); err != nil {
		return fmt.Errorf("failed to encrypt file: %w", err)
	}

	return nil
}
