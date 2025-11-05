// File : internal/backup/backup_checksum_verifier.go
// Deskripsi : Verifikasi checksum untuk backup files (decrypt + decompress + calculate)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package backup

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/encrypt"
	"strings"
)

// VerifyBackupChecksum melakukan verifikasi checksum untuk backup file
// Prosesnya: decrypt → decompress → calculate checksum → compare
func (s *Service) VerifyBackupChecksum(backupFilePath string, expectedSHA256, expectedMD5 string, encryptionKey string, compressionType string) (*types.ChecksumVerificationResult, error) {
	result := &types.ChecksumVerificationResult{
		FilePath:       backupFilePath,
		ExpectedSHA256: expectedSHA256,
		ExpectedMD5:    expectedMD5,
	}

	// Buka file backup
	file, err := os.Open(backupFilePath)
	if err != nil {
		result.Error = fmt.Sprintf("gagal membuka file: %v", err)
		return result, fmt.Errorf("gagal membuka file backup: %w", err)
	}
	defer file.Close()

	// Setup reader chain: File → Decrypt → Decompress → Hash calculation
	var reader io.Reader = file

	// Layer 1: Decryption (jika enabled)
	if encryptionKey != "" {
		decryptReader, err := encrypt.NewDecryptingReader(reader, encryptionKey)
		if err != nil {
			result.Error = fmt.Sprintf("gagal setup decrypt reader: %v", err)
			return result, fmt.Errorf("gagal setup decrypt reader: %w", err)
		}
		reader = decryptReader
	}

	// Layer 2: Decompression (jika ada)
	if compressionType != "" && compressionType != string(compress.CompressionNone) {
		decompressReader, err := compress.NewDecompressingReader(reader, compress.CompressionType(compressionType))
		if err != nil {
			result.Error = fmt.Sprintf("gagal setup decompress reader: %v", err)
			return result, fmt.Errorf("gagal setup decompress reader: %w", err)
		}
		defer decompressReader.Close()
		reader = decompressReader
	}

	// Calculate checksums
	sha256Hash := sha256.New()
	md5Hash := md5.New()
	multiWriter := io.MultiWriter(sha256Hash, md5Hash)

	if _, err := io.Copy(multiWriter, reader); err != nil {
		result.Error = fmt.Sprintf("gagal membaca data: %v", err)
		return result, fmt.Errorf("gagal membaca dan hash data: %w", err)
	}

	// Get calculated hashes
	result.CalculatedSHA256 = hex.EncodeToString(sha256Hash.Sum(nil))
	result.CalculatedMD5 = hex.EncodeToString(md5Hash.Sum(nil))

	// Compare checksums
	result.SHA256Match = strings.EqualFold(result.CalculatedSHA256, expectedSHA256)
	result.MD5Match = strings.EqualFold(result.CalculatedMD5, expectedMD5)
	result.Success = result.SHA256Match && result.MD5Match

	if !result.Success {
		if !result.SHA256Match {
			result.Error = fmt.Sprintf("SHA256 mismatch: expected=%s, got=%s", expectedSHA256[:16]+"...", result.CalculatedSHA256[:16]+"...")
		}
		if !result.MD5Match {
			if result.Error != "" {
				result.Error += "; "
			}
			result.Error += fmt.Sprintf("MD5 mismatch: expected=%s, got=%s", expectedMD5[:16]+"...", result.CalculatedMD5[:16]+"...")
		}
	}

	return result, nil
}
