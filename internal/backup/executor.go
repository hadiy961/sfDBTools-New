// File : internal/backup/executor.go
// Deskripsi : Entry point dan executor logic untuk backup operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/backup/modes"
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

	// Cleanup old backups jika enabled (runtime options; default dari config)
	if s.BackupDBOptions != nil && s.BackupDBOptions.Cleanup.Enabled && ctx.Err() == nil {
		s.Log.Info("Menjalankan cleanup old backups setelah backup...")
		if err := cleanup.CleanupOldBackupsFromBackup(s.Config, s.Log); err != nil {
			s.Log.Warnf("Cleanup old backups gagal: %v", err)
		}
	}

	// Handle errors
	finalResult, err := s.handleBackupErrors(result)
	if err != nil {
		return &finalResult, err
	}

	return &finalResult, nil
}

// executeBackupByMode menjalankan backup sesuai mode yang dipilih
func (s *Service) executeBackupByMode(ctx context.Context, dbFiltered []string, backupMode string) types_backup.BackupResult {
	executor, err := modes.GetExecutor(backupMode, s)
	if err != nil {
		s.Log.Errorf("Gagal mendapatkan executor: %v", err)
		return types_backup.BackupResult{
			Errors: []string{fmt.Sprintf("gagal inisialisasi mode backup: %v", err)},
		}
	}

	return executor.Execute(ctx, dbFiltered)
}

// handleBackupErrors menangani error dari backup execution
func (s *Service) handleBackupErrors(result types_backup.BackupResult) (types_backup.BackupResult, error) {
	if len(result.Errors) > 0 || len(result.FailedDatabaseInfos) > 0 {
		errorMsg := "backup gagal"
		if len(result.Errors) > 0 {
			errorMsg = result.Errors[0]
			// Aggregate multiple errors untuk visibility
			if len(result.Errors) > 1 {
				errorMsg = fmt.Sprintf("%s (dan %d error lainnya)", result.Errors[0], len(result.Errors)-1)
			}
		}
		return result, fmt.Errorf("%s", errorMsg)
	}

	return result, nil
}
