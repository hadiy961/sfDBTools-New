// File : internal/backup/executor.go
// Deskripsi : Entry point dan executor logic untuk backup operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/cleanup"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/helper"
)

// ExecuteBackup melakukan proses backup database - entry point utama
func (s *Service) ExecuteBackup(ctx context.Context, sourceClient *database.Client, dbFiltered []string, backupMode string) (*types_backup.BackupResult, error) {
	// Simpan client ke service
	s.Client = sourceClient

	// Setup konfigurasi backup
	if err := s.SetupBackupExecution(); err != nil {
		return nil, fmt.Errorf("gagal setup backup execution: %w", err)
	}

	// Jalankan backup sesuai mode
	timer := helper.NewTimer()
	result := s.executeBackupByMode(ctx, dbFiltered, backupMode)
	result.TotalTimeTaken = timer.Elapsed()

	// Cleanup old backups jika enabled
	if s.Config.Backup.Cleanup.Enabled {
		s.Log.Info("Menjalankan cleanup old backups setelah backup...")
		if err := cleanup.CleanupOldBackupsFromBackup(s.Config, s.Log); err != nil {
			s.Log.Warnf("Cleanup old backups gagal: %v", err)
		}
	}

	// Handle errors
	finalResult, err := s.handleBackupErrors(ctx, result)
	if err != nil {
		return &finalResult, err
	}

	return &finalResult, nil
}

// executeBackupByMode menjalankan backup sesuai mode yang dipilih
func (s *Service) executeBackupByMode(ctx context.Context, dbFiltered []string, backupMode string) types_backup.BackupResult {
	switch backupMode {
	case "separate", "separated":
		return s.ExecuteBackupSeparated(ctx, dbFiltered)
	case "single", "primary", "secondary":
		return s.ExecuteBackupSingle(ctx, dbFiltered)
	default:
		return s.ExecuteBackupCombined(ctx, dbFiltered)
	}
}

// handleBackupErrors menangani error dari backup execution
func (s *Service) handleBackupErrors(ctx context.Context, result types_backup.BackupResult) (types_backup.BackupResult, error) {
	if len(result.Errors) > 0 || len(result.FailedDatabaseInfos) > 0 {
		errorMsg := "backup gagal"
		if len(result.Errors) > 0 {
			errorMsg = fmt.Sprintf("backup gagal: %s", result.Errors[0])
		}
		return result, fmt.Errorf("%s", errorMsg)
	}

	return result, nil
}
