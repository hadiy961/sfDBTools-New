// File : internal/crypto/file/decrypt.go
// Deskripsi : File decryption convenience functions
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

// Decrypt decrypts a file using streaming decryption.
//
// Parameters:
//   - inputPath: path to encrypted file
//   - outputPath: path to decrypted output file
//   - passphrase: decryption password/key
//
// Returns error if decryption fails.
//
// Example:
//
//	err := file.Decrypt("data.txt.enc", "data.txt", []byte("my-password"))
func Decrypt(inputPath, outputPath string, passphrase []byte) error {
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

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Stream copy with decryption
	if _, err := io.Copy(outputFile, decReader); err != nil {
		return fmt.Errorf("failed to decrypt file: %w", err)
	}

	return nil
}
