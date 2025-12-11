// File : internal/backup/metadata/backup_metadata.go
// Deskripsi : Metadata generation dan penyimpanan untuk backup operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/global"
	"time"
)

// GenerateBackupMetadata membuat metadata backup dari config
func GenerateBackupMetadata(cfg types_backup.MetadataConfig) *types_backup.BackupMetadata {
	meta := &types_backup.BackupMetadata{
		BackupFile:      cfg.BackupFile,
		BackupType:      cfg.BackupType,
		DatabaseNames:   cfg.DatabaseNames,
		Hostname:        cfg.Hostname,
		BackupStartTime: cfg.StartTime,
		BackupEndTime:   cfg.EndTime,
		BackupDuration:  global.FormatDuration(cfg.Duration),
		FileSize:        cfg.FileSize,
		FileSizeHuman:   global.FormatFileSize(cfg.FileSize),
		Compressed:      cfg.Compressed,
		CompressionType: cfg.CompressionType,
		Encrypted:       cfg.Encrypted,
		BackupStatus:    cfg.BackupStatus,
		Warnings:        cfg.Warnings,
		GeneratedBy:     "sfDBTools",
		GeneratedAt:     time.Now(),
		// Additional files
		UserGrantsFile: cfg.UserGrantsFile,
		// Version information
		MysqldumpVersion: cfg.MysqldumpVersion,
		MariaDBVersion:   cfg.MariaDBVersion,
	}

	// GTID info dan replication info hanya disimpan untuk mode non-separated (combined, single, dll)
	// Untuk mode separated/multi-file, informasi ini tidak relevan karena setiap database dibackup terpisah
	if cfg.BackupType != "separated" {
		meta.GTIDInfo = cfg.GTIDInfo
		meta.ReplicationUser = cfg.ReplicationUser
		meta.ReplicationPassword = cfg.ReplicationPassword
		meta.SourceHost = cfg.SourceHost
		meta.SourcePort = cfg.SourcePort
	}

	if cfg.StderrOutput != "" {
		meta.Warnings = append(meta.Warnings, cfg.StderrOutput)
	}

	return meta
}

// SaveBackupMetadata menyimpan metadata ke file dengan atomic write
// Menggunakan temporary file + rename pattern untuk atomicity
func SaveBackupMetadata(meta *types_backup.BackupMetadata, logger applog.Logger) (string, error) {
	if meta == nil {
		return "", fmt.Errorf("metadata is nil")
	}

	manifestPath := meta.BackupFile + ".meta.json"
	tmpManifest := manifestPath + ".tmp"

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

	logger.Debugf("Metadata berhasil disimpan: %s", manifestPath)
	return manifestPath, nil
}

// TrySaveBackupMetadata adalah wrapper non-fatal untuk SaveBackupMetadata
// Jika gagal, log warning tapi jangan return error
// Berguna untuk metadata yang optional (tidak block backup jika gagal)
func TrySaveBackupMetadata(meta *types_backup.BackupMetadata, logger applog.Logger) string {
	if meta == nil {
		return ""
	}

	path, err := SaveBackupMetadata(meta, logger)
	if err != nil {
		logger.Warnf("Gagal menyimpan metadata backup: %v", err)
		return ""
	}

	return path
}

// UpdateMetadataUserGrantsFile membaca metadata yang ada, update UserGrantsFile field, dan save kembali
func UpdateMetadataUserGrantsFile(backupFilePath string, userGrantsPath string, logger applog.Logger) error {
	manifestPath := backupFilePath + ".meta.json"

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

// UpdateMetadataDatabaseNames membaca metadata file, update DatabaseNames, dan save kembali
func UpdateMetadataDatabaseNames(backupFilePath string, databaseNames []string, logger applog.Logger) error {
	metadataPath := backupFilePath + ".meta.json"

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

	// Update DatabaseNames
	meta.DatabaseNames = databaseNames
	logger.Debugf("Update metadata DatabaseNames: %v", databaseNames)

	// Save kembali
	_, err = SaveBackupMetadata(&meta, logger)
	return err
}

// UpdateMetadataWithDatabaseDetails update metadata dengan detail lengkap per database
func UpdateMetadataWithDatabaseDetails(backupFilePath string, databaseNames []string, backupInfos []types.DatabaseBackupInfo, logger applog.Logger) error {
	metadataPath := backupFilePath + ".meta.json"

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

// DatabaseBackupInfoBuilder membantu construct DatabaseBackupInfo dengan konsisten
type DatabaseBackupInfoBuilder struct {
	DatabaseName string
	OutputFile   string
	FileSize     int64
	Duration     time.Duration
	Status       string
	Warnings     string
	StartTime    time.Time
	EndTime      time.Time
	ManifestFile string
}

// Build membuat DatabaseBackupInfo dari builder
func (b *DatabaseBackupInfoBuilder) Build() types.DatabaseBackupInfo {
	// Calculate throughput
	var throughputMBs float64
	if b.Duration.Seconds() > 0 {
		throughputMBs = float64(b.FileSize) / (1024.0 * 1024.0) / b.Duration.Seconds()
	}

	// Generate backup ID
	backupID := fmt.Sprintf("bk-%d", time.Now().UnixNano())

	return types.DatabaseBackupInfo{
		DatabaseName:   b.DatabaseName,
		OutputFile:     b.OutputFile,
		FileSize:       b.FileSize,
		FileSizeHuman:  global.FormatFileSize(b.FileSize),
		Duration:       global.FormatDuration(b.Duration),
		Status:         b.Status,
		Warnings:       b.Warnings,
		BackupID:       backupID,
		StartTime:      b.StartTime,
		EndTime:        b.EndTime,
		ThroughputMBps: throughputMBs,
		ManifestFile:   b.ManifestFile,
	}
}
