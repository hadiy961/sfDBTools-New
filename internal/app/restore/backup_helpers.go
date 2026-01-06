// File : internal/restore/backup_helpers.go
// Deskripsi : Helper functions untuk backup pre-restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 19 Desember 2025
// Last Modified : 6 Januari 2026
package restore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfdbtools/internal/app/backup"
	"sfdbtools/internal/app/backup/model/types_backup"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/domain"
	"sfdbtools/pkg/compress"
	"sfdbtools/pkg/consts"
	"time"
)

// BackupDatabasesSingleFileIfNeeded membuat backup gabungan (single-file/combined)
// untuk sekumpulan database sebelum restore (dipakai oleh restore all).
// Konsepnya mengikuti backup filter --mode single-file.
func (s *Service) BackupDatabasesSingleFileIfNeeded(ctx context.Context, dbNames []string, skipBackup bool, backupOpts *restoremodel.RestoreBackupOptions) (string, error) {
	if skipBackup {
		return "", nil
	}
	if len(dbNames) == 0 {
		return "", nil
	}

	backupFile, err := s.backupDatabaseGeneric(ctx, consts.ModeCombined, "", dbNames, backupOpts)
	if err != nil {
		return "", err
	}
	return backupFile, nil
}

// backupDatabaseGeneric melakukan backup database menggunakan backup service (generic version)
// mode: "single" atau "all"
// dbName: nama database (untuk single mode)
// dbList: list of databases (untuk all mode)
func (s *Service) backupDatabaseGeneric(ctx context.Context, mode string, dbName string, dbList []string, backupOpts *restoremodel.RestoreBackupOptions) (string, error) {
	// Defensive defaulting: backupOpts can be nil if caller didn't run restore setup.
	// Use config defaults in that case.
	if backupOpts == nil {
		backupOpts = &restoremodel.RestoreBackupOptions{}
	}
	if backupOpts.Compression.Type == "" {
		backupOpts.Compression = domain.CompressionOptions{
			Enabled: s.Config.Backup.Compression.Enabled,
			Type:    s.Config.Backup.Compression.Type,
			Level:   s.Config.Backup.Compression.Level,
		}
	}
	if backupOpts.Encryption.Key == "" {
		// For pre-restore backups, default to backup encryption settings.
		backupOpts.Encryption = domain.EncryptionOptions{
			Enabled: s.Config.Backup.Encryption.Enabled,
			Key:     s.Config.Backup.Encryption.Key,
		}
	}

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
	switch mode {
	case consts.ModeAll:
		filename = fmt.Sprintf("all_%s_%s_pre_restore", timestamp, hostname)
	case consts.ModeCombined:
		filename = fmt.Sprintf("combined_%d_%s_%s_pre_restore", len(dbList), timestamp, hostname)
	default:
		filename = fmt.Sprintf("%s_%s_%s_pre_restore", dbName, timestamp, hostname)
	}

	fullFilename := filename + consts.ExtSQL

	if backupOpts.Compression.Enabled {
		ext := compress.GetFileExtension(compress.CompressionType(backupOpts.Compression.Type))
		fullFilename += ext
	}
	if backupOpts.Encryption.Enabled {
		fullFilename += consts.ExtEnc
	}

	outputPath := filepath.Join(outputDir, fullFilename)

	// Prepare backup options
	ticket := ""
	// gunakan ticket restore (jika tersedia) untuk metadata/audit
	if s.RestoreAllOpts != nil && s.RestoreAllOpts.Ticket != "" {
		ticket = s.RestoreAllOpts.Ticket
	} else if s.RestorePrimaryOpts != nil && s.RestorePrimaryOpts.Ticket != "" {
		ticket = s.RestorePrimaryOpts.Ticket
	} else if s.RestoreOpts != nil && s.RestoreOpts.Ticket != "" {
		ticket = s.RestoreOpts.Ticket
	} else if s.RestoreSelOpts != nil && s.RestoreSelOpts.Ticket != "" {
		ticket = s.RestoreSelOpts.Ticket
	}

	backupOptions := &types_backup.BackupDBOptions{
		Profile:   *s.Profile,
		OutputDir: outputDir,
		Mode:      mode,
		File: types_backup.BackupFileInfo{
			Filename: filename,
		},
		Compression: domain.CompressionOptions{
			Enabled: backupOpts.Compression.Enabled,
			Type:    backupOpts.Compression.Type,
			Level:   backupOpts.Compression.Level,
		},
		Encryption: domain.EncryptionOptions{
			Enabled: backupOpts.Encryption.Enabled,
			Key:     backupOpts.Encryption.Key,
		},
		Filter: domain.FilterOptions{},
		Ticket: ticket,
	}

	// Set filter based on mode
	switch mode {
	case consts.ModeSingle:
		backupOptions.Filter.IncludeDatabases = []string{dbName}
	case consts.ModeCombined:
		backupOptions.Filter.IncludeDatabases = dbList
		backupOptions.Filter.IsFilterCommand = true
	}

	// Create backup service and execute
	backupSvc := backup.NewBackupService(s.Log, s.Config, backupOptions)
	// Reuse current target connection for grants/GTID helpers.
	backupSvc.Client = s.TargetClient

	// Capture GTID (best-effort; follow behavior in combined mode executor)
	if capErr := backupSvc.CaptureAndSaveGTID(ctx, outputPath); capErr != nil {
		s.Log.Warnf("GTID handling error: %v", capErr)
	}

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
		IsMultiDB:    mode == consts.ModeAll || mode == consts.ModeCombined,
	}

	_, err := backupSvc.ExecuteAndBuildBackup(ctx, backupConfig)
	if err != nil {
		return "", fmt.Errorf("gagal backup database: %w", err)
	}

	// Export user grants untuk combined (single-file) agar konsisten dengan backup filter --mode single-file
	if mode == consts.ModeCombined {
		actualUserGrantsPath := backupSvc.ExportUserGrantsIfNeeded(ctx, outputPath, dbList)
		backupSvc.UpdateMetadataUserGrantsPath(outputPath, actualUserGrantsPath)
	}

	return outputPath, nil
}

// BackupTargetDatabase melakukan backup database target menggunakan backup service
func (s *Service) BackupTargetDatabase(ctx context.Context, dbName string, backupOpts *restoremodel.RestoreBackupOptions) (string, error) {
	return s.backupDatabaseGeneric(ctx, consts.ModeSingle, dbName, []string{dbName}, backupOpts)
}

// BackupAllDatabases melakukan backup semua database sebelum restore all
func (s *Service) BackupAllDatabases(ctx context.Context, backupOpts *restoremodel.RestoreBackupOptions) (string, error) {
	// Get DB list for count
	dbList, err := s.TargetClient.GetDatabaseList(ctx)
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan list database: %w", err)
	}

	return s.backupDatabaseGeneric(ctx, consts.ModeAll, "", dbList, backupOpts)
}
