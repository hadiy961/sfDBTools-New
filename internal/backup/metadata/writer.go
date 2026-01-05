// File : internal/backup/metadata/writer.go
// Deskripsi : Metadata file writing operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-22
// Last Modified :  2026-01-05
package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"sfDBTools/internal/services/log"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/consts"
)

// SaveBackupMetadata menyimpan metadata ke file dengan atomic write
// Menggunakan temporary file + rename pattern untuk atomicity
func SaveBackupMetadata(meta *types_backup.BackupMetadata, logger applog.Logger) (string, error) {
	if meta == nil {
		return "", fmt.Errorf("metadata is nil")
	}

	manifestPath := meta.BackupFile + consts.ExtMetaJSON
	tmpManifest := manifestPath + consts.ExtTmp

	// Marshal metadata to JSON
	manifestBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		logger.Errorf("Gagal membuat manifest JSON: %v", err)
		return "", fmt.Errorf("marshal metadata error: %w", err)
	}

	// Write to temporary file
	if err := os.WriteFile(tmpManifest, manifestBytes, 0644); err != nil {
		logger.Errorf("Gagal menulis file manifest sementara: %v", err)
		os.Remove(tmpManifest) // Clean up temp file
		return "", fmt.Errorf("write temporary file error: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpManifest, manifestPath); err != nil {
		logger.Errorf("Gagal merename manifest sementara ke target: %v", err)
		os.Remove(tmpManifest) // Clean up temp file
		return "", fmt.Errorf("rename file error: %w", err)
	}

	// logger.Debugf("Metadata berhasil disimpan: %s", manifestPath)
	return manifestPath, nil
}

// TrySaveBackupMetadata adalah wrapper non-fatal untuk SaveBackupMetadata
// Jika gagal, log warning tapi jangan return error
// Berguna untuk metadata yang optional (tidak block backup jika gagal)
func TrySaveBackupMetadata(meta *types_backup.BackupMetadata, logger applog.Logger) string {
	if meta == nil {
		logger.Debug("Skip save metadata: meta is nil")
		return ""
	}

	path, err := SaveBackupMetadata(meta, logger)
	if err != nil {
		logger.Warnf("Gagal menyimpan metadata backup: %v", err)
		return ""
	}

	return path
}
