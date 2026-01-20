// File : internal/backup/metadata/writer.go
// Deskripsi : Metadata file writing operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-22
// Last Modified : 2026-01-20
package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"sfdbtools/internal/app/backup/model"
	"sfdbtools/internal/app/backup/model/types_backup"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
	"strconv"
)

// SaveBackupMetadata menyimpan metadata ke file dengan atomic write
// Menggunakan temporary file + rename pattern untuk atomicity
// permissions: string format "0600" atau "0644" (default: 0600 jika kosong/invalid)
func SaveBackupMetadata(meta *types_backup.BackupMetadata, permissions string, logger applog.Logger) (string, error) {
	if meta == nil {
		return "", fmt.Errorf("WriteMetadataToFile: %w", model.ErrMetadataIsNil)
	}

	manifestPath := meta.BackupFile + consts.ExtMetaJSON
	tmpManifest := manifestPath + consts.ExtTmp

	// Parse permissions string to os.FileMode
	perm := parseFilePermissions(permissions, logger)

	// Marshal metadata to JSON
	manifestBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		logger.Errorf("Gagal membuat manifest JSON: %v", err)
		return "", fmt.Errorf("marshal metadata error: %w", err)
	}

	// Write to temporary file
	if err := os.WriteFile(tmpManifest, manifestBytes, perm); err != nil {
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
func TrySaveBackupMetadata(meta *types_backup.BackupMetadata, permissions string, logger applog.Logger) string {
	if meta == nil {
		logger.Debug("Skip save metadata: meta is nil")
		return ""
	}

	path, err := SaveBackupMetadata(meta, permissions, logger)
	if err != nil {
		logger.Warnf("Gagal menyimpan metadata backup: %v", err)
		return ""
	}

	return path
}

// parseFilePermissions mengkonversi string permissions (e.g., "0600") ke os.FileMode
// Jika parsing gagal atau permissions kosong, return default 0600 (lebih restrictive)
func parseFilePermissions(permStr string, logger applog.Logger) os.FileMode {
	const defaultPerm = 0600

	if permStr == "" {
		return defaultPerm
	}

	// Parse octal string to uint32
	perm, err := strconv.ParseUint(permStr, 8, 32)
	if err != nil {
		logger.Warnf("Invalid metadata_permissions '%s', using default 0600: %v", permStr, err)
		return defaultPerm
	}

	return os.FileMode(perm)
}
