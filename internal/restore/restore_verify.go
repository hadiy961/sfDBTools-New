// File : internal/restore/restore_verify.go
// Deskripsi : Verifikasi backup file sebelum restore (checksum, metadata)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package restore

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/helper"
	"strings"
	"time"
)

// verifyBackupFile melakukan verifikasi backup file sebelum restore
func (s *Service) verifyBackupFile(ctx context.Context) error {
	sourceFile := s.RestoreOptions.SourceFile

	s.Log.Info("Reading backup file for verification...")

	// Baca file info
	fileInfo, err := os.Stat(sourceFile)
	if err != nil {
		return fmt.Errorf("gagal stat backup file: %w", err)
	}

	// Detect encryption dan compression dari extension
	isEncrypted := strings.HasSuffix(sourceFile, ".enc")
	compressionType := detectCompressionType(sourceFile)

	verifyInfo := &types.RestoreVerificationInfo{
		BackupFile:       sourceFile,
		FileSize:         fileInfo.Size(),
		Encrypted:        isEncrypted,
		Compressed:       compressionType != "",
		CompressionType:  compressionType,
		VerificationTime: time.Now(),
	}

	// Cari metadata file (.meta.json)
	// Format: backup_file.gz.enc -> backup_file.gz.enc.meta.json
	metadataFile := sourceFile + ".meta.json"

	s.Log.Debugf("Looking for metadata file: %s", metadataFile)

	// Load metadata jika ada
	var backupMetadata *types.BackupMetadata
	if _, err := os.Stat(metadataFile); err == nil {
		s.Log.Infof("✓ Loading metadata from: %s", metadataFile)
		metadata, err := s.loadBackupMetadata(metadataFile)
		if err != nil {
			s.Log.Warnf("Gagal load metadata: %v", err)
		} else {
			backupMetadata = metadata
			// Extract expected checksums from metadata
			for _, checksum := range metadata.Checksums {
				if checksum.Algorithm == "sha256" {
					verifyInfo.ExpectedSHA256 = checksum.Hash
				} else if checksum.Algorithm == "md5" {
					verifyInfo.ExpectedMD5 = checksum.Hash
				}
			}
		}
	} else {
		s.Log.Warn("Metadata file tidak ditemukan, skip checksum verification")
	}

	// Verify checksum jika ada metadata
	if verifyInfo.ExpectedSHA256 != "" || verifyInfo.ExpectedMD5 != "" {
		s.Log.Info("Calculating checksums dari backup file...")

		encryptionKey := s.RestoreOptions.EncryptionKey
		if encryptionKey == "" && isEncrypted {
			encryptionKey = helper.GetEnvOrDefault("SFDB_BACKUP_ENCRYPTION_KEY", "")
			if encryptionKey == "" {
				return fmt.Errorf("encryption key required untuk verify encrypted backup")
			}
		}

		calculatedSHA256, calculatedMD5, err := s.calculateBackupChecksum(sourceFile, encryptionKey, compressionType)
		if err != nil {
			verifyInfo.ErrorMessage = fmt.Sprintf("Gagal calculate checksum: %v", err)
			return fmt.Errorf("gagal calculate checksum: %w", err)
		}

		verifyInfo.CalculatedSHA256 = calculatedSHA256
		verifyInfo.CalculatedMD5 = calculatedMD5

		// Compare checksums
		sha256Match := strings.EqualFold(calculatedSHA256, verifyInfo.ExpectedSHA256)
		md5Match := strings.EqualFold(calculatedMD5, verifyInfo.ExpectedMD5)
		verifyInfo.ChecksumMatch = sha256Match && md5Match

		if !verifyInfo.ChecksumMatch {
			errMsg := "Checksum mismatch:"
			if !sha256Match {
				errMsg += fmt.Sprintf(" SHA256 (expected=%s, got=%s)",
					verifyInfo.ExpectedSHA256[:16]+"...", calculatedSHA256[:16]+"...")
			}
			if !md5Match {
				errMsg += fmt.Sprintf(" MD5 (expected=%s, got=%s)",
					verifyInfo.ExpectedMD5[:16]+"...", calculatedMD5[:16]+"...")
			}
			verifyInfo.ErrorMessage = errMsg
			return fmt.Errorf(errMsg)
		}

		s.Log.Info("✓ Checksum verification SUCCESS")
	}

	// Log verification info
	s.Log.Infof("Backup File: %s", verifyInfo.BackupFile)
	s.Log.Infof("File Size: %d bytes", verifyInfo.FileSize)
	s.Log.Infof("Encrypted: %v", verifyInfo.Encrypted)
	s.Log.Infof("Compressed: %v (%s)", verifyInfo.Compressed, verifyInfo.CompressionType)
	if backupMetadata != nil {
		s.Log.Infof("Databases in backup: %v", backupMetadata.DatabaseNames)
	}

	return nil
}

// loadBackupMetadata load metadata dari file .meta.json
func (s *Service) loadBackupMetadata(metadataFile string) (*types.BackupMetadata, error) {
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, err
	}

	var metadata types.BackupMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// calculateBackupChecksum calculate SHA256 dan MD5 dari backup file
func (s *Service) calculateBackupChecksum(backupFile, encryptionKey, compressionType string) (string, string, error) {
	file, err := os.Open(backupFile)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	var reader io.Reader = file

	// Decrypt jika encrypted
	if encryptionKey != "" {
		decryptReader, err := encrypt.NewDecryptingReader(reader, encryptionKey)
		if err != nil {
			return "", "", fmt.Errorf("gagal create decrypt reader: %w", err)
		}
		reader = decryptReader
	}

	// Decompress jika compressed
	if compressionType != "" {
		decompressReader, err := compress.NewDecompressingReader(reader, compress.CompressionType(compressionType))
		if err != nil {
			return "", "", fmt.Errorf("gagal create decompress reader: %w", err)
		}
		defer decompressReader.Close()
		reader = decompressReader
	}

	// Calculate checksums
	sha256Hash := sha256.New()
	md5Hash := md5.New()
	multiWriter := io.MultiWriter(sha256Hash, md5Hash)

	if _, err := io.Copy(multiWriter, reader); err != nil {
		return "", "", fmt.Errorf("gagal read and hash data: %w", err)
	}

	return hex.EncodeToString(sha256Hash.Sum(nil)), hex.EncodeToString(md5Hash.Sum(nil)), nil
}

// detectCompressionType detect compression type dari file extension
func detectCompressionType(filename string) string {
	// Remove .enc extension first if exists
	name := strings.TrimSuffix(filename, ".enc")
	name = strings.ToLower(name)

	if strings.HasSuffix(name, ".gz") {
		return "gzip"
	} else if strings.HasSuffix(name, ".zst") {
		return "zstd"
	} else if strings.HasSuffix(name, ".xz") {
		return "xz"
	} else if strings.HasSuffix(name, ".zlib") {
		return "zlib"
	}

	return ""
}
