// File : internal/restore/restore_multi.go
// Deskripsi : Restore multiple databases dari multiple backup files dalam direktori
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-10
// Last Modified : 2025-11-10

package restore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/profilehelper"
	"sfDBTools/pkg/servicehelper"
	"sfDBTools/pkg/ui"
	"time"
)

// BackupFileInfo menyimpan informasi tentang file backup yang ditemukan
type BackupFileInfo struct {
	FilePath     string
	DatabaseName string
	ModTime      time.Time
	FileSize     int64
}

// executeRestoreMulti melakukan restore multiple databases dari direktori backup files
func (s *Service) executeRestoreMulti(ctx context.Context) (types.RestoreResult, error) {
	defer servicehelper.TrackProgress(s)()

	// Mark restore as in progress untuk graceful shutdown
	s.SetRestoreInProgress(true)
	defer s.SetRestoreInProgress(false)

	sourceDir := s.RestoreOptions.SourceFile

	ui.PrintSubHeader("Scanning Direktori Backup")
	s.Log.Infof("Scanning direktori untuk backup files: %s", sourceDir)

	// Scan direktori untuk mendapatkan semua backup files
	backupFiles, err := s.scanBackupFiles(sourceDir)
	if err != nil {
		return types.RestoreResult{}, fmt.Errorf("gagal scan backup files: %w", err)
	}

	if len(backupFiles) == 0 {
		return types.RestoreResult{}, fmt.Errorf("tidak ada backup file ditemukan di direktori: %s", sourceDir)
	}

	s.Log.Infof("Ditemukan %d backup files", len(backupFiles))

	// Group files berdasarkan database name dan ambil yang terbaru untuk setiap database
	latestFiles := s.SelectLatestBackupFiles(backupFiles)

	s.Log.Infof("Total %d database unik yang akan di-restore", len(latestFiles))

	// Check if dry run
	if s.RestoreOptions.DryRun {
		ui.PrintSubHeader("Mode Simulasi (Dry Run)")
		s.Log.Info("[DRY RUN] Restore tidak dijalankan (mode simulasi)")
		builder := NewRestoreResultBuilder()
		builder.SetTotalDatabases(len(latestFiles))

		for _, fileInfo := range latestFiles {
			info := buildSkippedRestoreInfo(
				fileInfo.DatabaseName,
				fileInfo.FilePath,
				fileInfo.DatabaseName,
				fileInfo.FileSize,
				global.FormatFileSize(fileInfo.FileSize),
			)
			builder.AddSkipped(info)
		}
		return builder.Build(), nil
	}

	builder := NewRestoreResultBuilder()
	builder.SetTotalDatabases(len(latestFiles))

	// Jika opsi SkipBackup == false maka buat pre-backup untuk semua database
	// sebelum melakukan restore. Fungsi ExecuteMultiPreBackup sudah tersedia
	// namun sebelumnya tidak pernah dipanggil.
	if !s.RestoreOptions.SkipBackup {
		ui.PrintSubHeader("Membuat Safety Backup untuk Semua Database")
		s.Log.Info("Creating safety backups sebelum restore...")
		preBackupDir, err := s.ExecuteMultiPreBackup(ctx, latestFiles)
		if err != nil {
			s.Log.Warnf("Gagal membuat pre-backup untuk multiple restore: %v", err)
		} else if preBackupDir != "" {
			builder.SetPreBackupFile(preBackupDir)
			s.Log.Infof("Pre-backup directory created: %s", preBackupDir)
		}
	}

	// Execute restore untuk setiap database
	ui.PrintSubHeader("Melakukan Restore Multiple Databases")
	s.Log.Infof("Starting restore untuk %d database...", len(latestFiles))

	for i, fileInfo := range latestFiles {
		// Check context cancellation untuk graceful shutdown
		select {
		case <-ctx.Done():
			s.Log.Warn("Restore dibatalkan, menghentikan proses...")
			return builder.Build(), ctx.Err()
		default:
		}

		s.Log.Infof("[%d/%d] Restoring database: %s from %s",
			i+1, len(latestFiles), fileInfo.DatabaseName, filepath.Base(fileInfo.FilePath))

		// Prepare database: drop (jika flag aktif) → create (jika tidak ada)
		// prepareDatabaseForRestore sudah melakukan ensureValidConnection di dalamnya
		prepResult, err := s.prepareDatabaseForRestore(ctx, fileInfo.DatabaseName, s.RestoreOptions.DropTarget)
		if err != nil {
			s.Log.Errorf("Gagal prepare database %s: %v", fileInfo.DatabaseName, err)

			failedInfo := buildFailedRestoreInfo(
				fileInfo.DatabaseName,
				fileInfo.FilePath,
				fileInfo.DatabaseName,
				prepResult.ErrorMessage,
				fileInfo.FileSize,
				global.FormatFileSize(fileInfo.FileSize),
			)
			builder.AddFailureWithPrefix(fileInfo.DatabaseName, failedInfo, 0, err, fmt.Sprintf("Database %s", fileInfo.DatabaseName))
			continue
		}

		// Restore database menggunakan fungsi restore single yang sudah ada
		dbTimer := helper.NewTimer()
		restoreInfo, err := s.restoreSingleDatabase(ctx, fileInfo.FilePath, fileInfo.DatabaseName, fileInfo.DatabaseName)
		dbDuration := dbTimer.Elapsed()

		// Force reconnect setelah restore untuk cleanup stale connections
		// dan prepare fresh connection untuk database berikutnya
		if i < len(latestFiles)-1 { // Skip di iterasi terakhir
			s.Log.Debug("Refreshing database connection untuk database berikutnya...")
			if s.Client != nil {
				s.Client.Close() // Close existing connection
			}
			client, connErr := profilehelper.ConnectWithProfile(s.TargetProfile, "mysql")
			if connErr != nil {
				s.Log.Warnf("Gagal refresh connection: %v", connErr)
			} else {
				s.Client = client
				s.Log.Debug("✓ Connection refreshed")
			}
		}

		if err != nil {
			builder.AddFailureWithPrefix(fileInfo.DatabaseName, restoreInfo, dbDuration, err, fmt.Sprintf("Database %s", fileInfo.DatabaseName))
		} else {
			// Jika ada warning (dari force mode), log ke result juga
			if restoreInfo.Warnings != "" {
				s.Log.Warnf("Restore success dengan warning untuk %s: %s", fileInfo.DatabaseName, restoreInfo.Warnings)
				builder.AddSuccessWithWarning(restoreInfo, dbDuration, fmt.Sprintf("WARNING for %s: %s", fileInfo.DatabaseName, restoreInfo.Warnings))
			} else {
				builder.AddSuccess(restoreInfo, dbDuration)
			}
		}
	}

	return builder.Build(), nil
}

// executeMultiPreBackup membuat safety backup untuk semua databases yang akan di-restore
func (s *Service) ExecuteMultiPreBackup(ctx context.Context, filesInfo []BackupFileInfo) (string, error) {
	// Create backup directory dengan timestamp
	backupDir := fmt.Sprintf("pre_restore_backup_%s", time.Now().Format("20060102_150405"))

	// Use absolute path in temp or configured backup dir
	backupBasePath := filepath.Join(s.Config.Backup.Output.BaseDirectory, backupDir)
	if err := os.MkdirAll(backupBasePath, 0755); err != nil {
		return "", fmt.Errorf("gagal create backup directory: %w", err)
	}

	s.Log.Infof("Pre-backup directory: %s", backupBasePath)

	// Collect database existence info sekali saja untuk semua database
	// untuk avoid duplicate check saat loop backup
	dbExistenceMap := make(map[string]bool)
	for _, fileInfo := range filesInfo {
		s.Log.Debugf("Checking database existence: %s", fileInfo.DatabaseName)
		exists, err := s.Client.DatabaseExists(ctx, fileInfo.DatabaseName)
		if err != nil {
			s.Log.Warnf("Gagal check keberadaan database %s: %v, anggap tidak ada", fileInfo.DatabaseName, err)
			dbExistenceMap[fileInfo.DatabaseName] = false
		} else {
			dbExistenceMap[fileInfo.DatabaseName] = exists
		}
	}

	for i, fileInfo := range filesInfo {
		// Check context cancellation untuk graceful shutdown
		select {
		case <-ctx.Done():
			s.Log.Warn("Pre-backup dibatalkan, menghentikan proses...")
			return backupBasePath, ctx.Err()
		default:
		}

		// Check database existence dari map (sudah di-check sebelumnya)
		exists := dbExistenceMap[fileInfo.DatabaseName]

		if !exists {
			s.Log.Infof("[%d/%d] Database %s tidak ada, skip pre-backup",
				i+1, len(filesInfo), fileInfo.DatabaseName)
			continue
		}

		s.Log.Infof("[%d/%d] Creating pre-backup untuk database: %s",
			i+1, len(filesInfo), fileInfo.DatabaseName)

		preBackupFile, err := s.executePreBackup(ctx, fileInfo.DatabaseName)
		if err != nil {
			s.Log.Warnf("Gagal create pre-backup untuk %s: %v", fileInfo.DatabaseName, err)
			// Continue dengan restore meskipun pre-backup gagal (warning only)
			continue
		}

		// Move pre-backup file ke backup directory
		destFile := filepath.Join(backupBasePath, filepath.Base(preBackupFile))
		if err := os.Rename(preBackupFile, destFile); err != nil {
			s.Log.Warnf("Gagal move pre-backup file untuk %s: %v", fileInfo.DatabaseName, err)
		} else {
			s.Log.Infof("✓ Pre-backup saved: %s", destFile)
		}
	}

	return backupBasePath, nil
}
