// File : internal/backup/metadata/updater.go
// Deskripsi : Metadata update operations
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

// UpdateMetadataUserGrantsFile membaca metadata yang ada, update UserGrantsFile field, dan save kembali
func UpdateMetadataUserGrantsFile(backupFilePath string, userGrantsPath string, logger applog.Logger) error {
	manifestPath := backupFilePath + consts.ExtMetaJSON

	// Baca metadata yang ada
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("gagal membaca metadata: %w", err)
	}

	// Parse JSON
	var meta types_backup.BackupMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("gagal parse metadata: %w", err)
	}

	// Update UserGrantsFile - jika empty string, set ke "none"
	if userGrantsPath == "" {
		meta.UserGrantsFile = "none"
	} else {
		meta.UserGrantsFile = userGrantsPath
	}

	// Save kembali
	_, err = SaveBackupMetadata(&meta, logger)
	return err
}

// UpdateMetadataWithDatabaseDetails update metadata dengan detail lengkap per database
func UpdateMetadataWithDatabaseDetails(backupFilePath string, databaseNames []string, backupInfos []types_backup.DatabaseBackupInfo, logger applog.Logger) error {
	metadataPath := backupFilePath + consts.ExtMetaJSON

	// Baca metadata file
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		logger.Warnf("Gagal membaca metadata file %s: %v", metadataPath, err)
		return err
	}

	// Parse JSON
	var meta types_backup.BackupMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		logger.Warnf("Gagal parse metadata JSON %s: %v", metadataPath, err)
		return err
	}

	// Pastikan BackupFile ter-set (untuk SaveBackupMetadata nanti)
	if meta.BackupFile == "" {
		meta.BackupFile = backupFilePath
	}

	// Update DatabaseNames
	meta.DatabaseNames = databaseNames

	// Build DatabaseDetails dari backupInfos
	details := make([]types_backup.DatabaseBackupDetail, 0, len(backupInfos))
	for _, info := range backupInfos {
		details = append(details, types_backup.DatabaseBackupDetail{
			DatabaseName:  info.DatabaseName,
			BackupFile:    info.OutputFile,
			FileSizeBytes: info.FileSize,
			FileSizeHuman: info.FileSizeHuman,
		})
	}
	meta.DatabaseDetails = details

	logger.Debugf("Update metadata dengan %d database details", len(details))

	// Save kembali
	_, err = SaveBackupMetadata(&meta, logger)
	return err
}
