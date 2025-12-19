// File : internal/restore/backup_helpers.go
// Deskripsi : Helper functions untuk backup pre-restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-19
// Last Modified : 2025-12-19

package restore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/compress"
	"time"
)

// backupDatabaseGeneric melakukan backup database menggunakan backup service (generic version)
// mode: "single" atau "all"
// dbName: nama database (untuk single mode)
// dbList: list of databases (untuk all mode)
func (s *Service) backupDatabaseGeneric(ctx context.Context, mode string, dbName string, dbList []string, backupOpts *types.RestoreBackupOptions) (string, error) {
	// Determine output directory
	outputDir := ""
	if backupOpts != nil && backupOpts.OutputDir != "" {
		outputDir = backupOpts.OutputDir
	} else {
		outputDir = s.Config.Backup.Output.BaseDirectory
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat direktori output: %w", err)
	}

	// Generate filename
	timestamp := time.Now().Format("20060102_150405")
	hostname := s.Profile.DBInfo.HostName
	if hostname == "" {
		hostname = s.Profile.DBInfo.Host
	}

	var filename string
	if mode == "all" {
		filename = fmt.Sprintf("all-databases_%s_%s_pre_restore", timestamp, hostname)
	} else {
		filename = fmt.Sprintf("%s_%s_%s_pre_restore", dbName, timestamp, hostname)
	}

	fullFilename := filename + ".sql"

	if backupOpts.Compression.Enabled {
		ext := compress.GetFileExtension(compress.CompressionType(backupOpts.Compression.Type))
		fullFilename += ext
	}
	if backupOpts.Encryption.Enabled {
		fullFilename += ".enc"
	}

	outputPath := filepath.Join(outputDir, fullFilename)

	// Prepare backup options
	backupOptions := &types_backup.BackupDBOptions{
		Profile:   *s.Profile,
		OutputDir: outputDir,
		Mode:      mode,
		File: types_backup.BackupFileInfo{
			Filename: filename,
		},
		Compression: types_backup.CompressionOptions{
			Enabled: backupOpts.Compression.Enabled,
			Type:    backupOpts.Compression.Type,
			Level:   backupOpts.Compression.Level,
		},
		Encryption: types_backup.EncryptionOptions{
			Enabled: backupOpts.Encryption.Enabled,
			Key:     backupOpts.Encryption.Key,
		},
		Filter: types.FilterOptions{},
	}

	// Set filter based on mode
	if mode == "single" {
		backupOptions.Filter.IncludeDatabases = []string{dbName}
	}

	// Create backup service and execute
	backupSvc := backup.NewBackupService(s.Log, s.Config, backupOptions)

	// Prepare backup config
	totalDBFound := len(dbList)
	if totalDBFound == 0 {
		totalDBFound = 1 // Fallback for single mode
	}

	backupConfig := types_backup.BackupExecutionConfig{
		DBName:       dbName,
		DBList:       dbList,
		OutputPath:   outputPath,
		BackupType:   mode,
		TotalDBFound: totalDBFound,
		IsMultiDB:    mode == "all",
	}

	_, err := backupSvc.ExecuteAndBuildBackup(ctx, backupConfig)
	if err != nil {
		return "", fmt.Errorf("gagal backup database: %w", err)
	}

	return outputPath, nil
}

// BackupTargetDatabase melakukan backup database target menggunakan backup service
func (s *Service) BackupTargetDatabase(ctx context.Context, dbName string, backupOpts *types.RestoreBackupOptions) (string, error) {
	return s.backupDatabaseGeneric(ctx, "single", dbName, []string{dbName}, backupOpts)
}

// BackupAllDatabases melakukan backup semua database sebelum restore all
func (s *Service) BackupAllDatabases(ctx context.Context, backupOpts *types.RestoreBackupOptions) (string, error) {
	// Get DB list for count
	dbList, err := s.TargetClient.GetDatabaseList(ctx)
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan list database: %w", err)
	}

	return s.backupDatabaseGeneric(ctx, "all", "", dbList, backupOpts)
}
