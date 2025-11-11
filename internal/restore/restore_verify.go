// File : internal/restore/restore_verify.go
// Deskripsi : Verifikasi backup file sebelum restore (basic validation)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-11

package restore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/global"
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

	// Log verification info
	s.Log.Infof("Backup File: %s", verifyInfo.BackupFile)
	s.Log.Infof("File Size: %s ", global.FormatFileSize(verifyInfo.FileSize))
	s.Log.Infof("Encrypted: %v", verifyInfo.Encrypted)
	s.Log.Infof("Compressed: %v (%s)", verifyInfo.Compressed, verifyInfo.CompressionType)

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

// Verifikasi apakah database target sudah ada
func (s *Service) isTargetDatabaseExists(ctx context.Context, targetDB string) (bool, error) {
	// exists, err := database.CheckDatabaseExists(ctx, s.Client, dbName)
	// if err != nil {
	// 	return false, fmt.Errorf("gagal cek keberadaan database target: %w", err)
	// }

	// Check if target database exists, create if not
	s.Log.Infof("Checking target database existence: %s", targetDB)
	exists, err := s.Client.DatabaseExists(ctx, targetDB)
	if err != nil {
		return false, fmt.Errorf("gagal check database existence: %w", err)
	}
	if exists {
		s.Log.Infof("✓ Database %s sudah ada", targetDB)
	}

	return exists, nil
}
