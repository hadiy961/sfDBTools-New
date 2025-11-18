// File : internal/restore/restore_verify.go
// Deskripsi : Verifikasi backup file sebelum restore (basic validation)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-11

package restore

import (
	"context"
	"fmt"
	"os"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/compress"
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
	compressionType := compress.DetectCompressionTypeFromFile(sourceFile)

	verifyInfo := &types.RestoreVerificationInfo{
		BackupFile:       sourceFile,
		FileSize:         fileInfo.Size(),
		Encrypted:        isEncrypted,
		Compressed:       compressionType != compress.CompressionNone,
		CompressionType:  string(compressionType),
		VerificationTime: time.Now(),
	}

	// Log verification info
	s.Log.Infof("Backup File: %s", verifyInfo.BackupFile)
	s.Log.Infof("File Size: %s ", global.FormatFileSize(verifyInfo.FileSize))
	s.Log.Infof("Encrypted: %v", verifyInfo.Encrypted)
	s.Log.Infof("Compressed: %v (%s)", verifyInfo.Compressed, verifyInfo.CompressionType)

	return nil
}
