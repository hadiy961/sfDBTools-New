// File : internal/backup/backup_executor.go
// Deskripsi : Executor untuk proses backup database dengan cleanup otomatis
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-15
// Last Modified : 2025-11-05

package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
	"time"
)

// ExecuteBackup melakukan proses backup database
func (s *Service) ExecuteBackup(ctx context.Context, sourceClient *database.Client, dbFiltered []string, backupMode string) (*types.BackupResult, error) {

	// Simpan client ke service untuk digunakan di fungsi lain
	s.Client = sourceClient

	// 1. Setup konfigurasi backup
	err := s.SetupBackupExecution()
	if err != nil {
		return nil, fmt.Errorf("gagal setup backup execution: %w", err)
	}

	// 2. Cleanup old backups sebelum backup baru (jika enabled)
	if s.Config.Backup.Cleanup.Enabled {
		s.Log.Info("Cleanup old backups enabled, menjalankan cleanup sebelum backup...")
		if err := s.cleanupOldBackups(); err != nil {
			s.Log.Warnf("Cleanup old backups gagal: %v (backup akan tetap dilanjutkan)", err)
		}
	}

	// 3. Eksekusi backup sesuai mode
	startTime := time.Now()
	var result types.BackupResult

	ui.PrintSubHeader("Memulai Proses Backup")

	if backupMode == "separate" || backupMode == "separated" {
		result = s.ExecuteBackupSeparated(ctx, dbFiltered)
	} else {
		result = s.ExecuteBackupCombined(ctx, dbFiltered)
	}

	result.TotalTimeTaken = time.Since(startTime)

	// 3. Cek apakah ada error dalam result
	if len(result.Errors) > 0 || len(result.FailedDatabaseInfos) > 0 {
		// Jika ada error, kembalikan sebagai error
		return &types.BackupResult{
			TotalDatabases:      len(dbFiltered),
			SuccessfulBackups:   0,
			FailedBackups:       len(dbFiltered),
			FailedDatabases:     map[string]string{},
			BackupInfo:          result.BackupInfo,
			FailedDatabaseInfos: result.FailedDatabaseInfos,
			Errors:              result.Errors,
			TotalTimeTaken:      result.TotalTimeTaken,
		}, fmt.Errorf("backup gagal: %s", result.Errors[0])
	}

	// 4. Kembalikan hasil backup sukses
	return &types.BackupResult{
		TotalDatabases:    len(dbFiltered),
		SuccessfulBackups: len(dbFiltered),
		FailedBackups:     0,
		FailedDatabases:   map[string]string{},
		BackupInfo:        result.BackupInfo,
		TotalTimeTaken:    result.TotalTimeTaken,
	}, nil
}
