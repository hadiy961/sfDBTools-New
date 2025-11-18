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
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
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

	// 2. Setup max_statement_time untuk GLOBAL (set ke unlimited untuk backup jangka panjang)
	// GLOBAL scope agar mysqldump juga affected (mysqldump membuat koneksi terpisah)
	// OPTIMISASI: Skip jika tidak critical untuk performa (query ini bisa lambat)
	s.Log.Debug("Skipping GLOBAL max_statement_time setup untuk performa optimal")
	// Uncomment jika diperlukan untuk backup jangka sangat panjang (>1 jam):
	// restore, originalMaxStatementTime, err := database.WithGlobalMaxStatementTime(ctx, sourceClient, 0)
	// if err != nil {
	// 	s.Log.Warnf("Setup GLOBAL max_statement_time gagal: %v", err)
	// } else {
	// 	s.Log.Infof("Original GLOBAL max_statement_time: %f detik", originalMaxStatementTime)
	// 	defer func() {
	// 		if rerr := restore(context.Background()); rerr != nil {
	// 			s.Log.Warnf("Gagal mengembalikan GLOBAL max_statement_time: %v", rerr)
	// 		} else {
	// 			s.Log.Info("GLOBAL max_statement_time berhasil dikembalikan.")
	// 		}
	// 	}()
	// }

	// 3. Cleanup old backups SETELAH backup baru (jika enabled)
	// OPTIMISASI: Pindahkan cleanup ke AFTER backup untuk tidak block backup execution
	cleanupDeferred := false
	if s.Config.Backup.Cleanup.Enabled {
		s.Log.Debug("Cleanup akan dijalankan setelah backup selesai")
		cleanupDeferred = true
	}

	// 4. Eksekusi backup sesuai mode
	timer := helper.NewTimer()
	var result types.BackupResult

	ui.PrintSubHeader("Memulai Proses Backup")

	if backupMode == "separate" || backupMode == "separated" {
		result = s.ExecuteBackupSeparated(ctx, dbFiltered)
	} else {
		result = s.ExecuteBackupCombined(ctx, dbFiltered)
	}

	result.TotalTimeTaken = timer.Elapsed()

	// 5. Cleanup old backups SETELAH backup selesai (jika enabled)
	if cleanupDeferred {
		s.Log.Info("Menjalankan cleanup old backups setelah backup...")
		if err := s.cleanupOldBackups(); err != nil {
			s.Log.Warnf("Cleanup old backups gagal: %v", err)
		}
	}

	// 6. Cek apakah ada error dalam result
	if len(result.Errors) > 0 || len(result.FailedDatabaseInfos) > 0 {
		// Jika ada error, kembalikan sebagai error
		errorMsg := "backup gagal"
		if len(result.Errors) > 0 {
			errorMsg = fmt.Sprintf("backup gagal: %s", result.Errors[0])
		}
		return &types.BackupResult{
			TotalDatabases:      len(dbFiltered),
			SuccessfulBackups:   0,
			FailedBackups:       len(dbFiltered),
			FailedDatabases:     map[string]string{},
			BackupInfo:          result.BackupInfo,
			FailedDatabaseInfos: result.FailedDatabaseInfos,
			Errors:              result.Errors,
			TotalTimeTaken:      result.TotalTimeTaken,
		}, fmt.Errorf(errorMsg)
	}

	// 7. Kembalikan hasil backup sukses
	return &types.BackupResult{
		TotalDatabases:    len(dbFiltered),
		SuccessfulBackups: len(dbFiltered),
		FailedBackups:     0,
		FailedDatabases:   map[string]string{},
		BackupInfo:        result.BackupInfo,
		TotalTimeTaken:    result.TotalTimeTaken,
	}, nil
}
