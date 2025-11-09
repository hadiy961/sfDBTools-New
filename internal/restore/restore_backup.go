// File : internal/restore/restore_backup.go
// Deskripsi : Pre-backup functionality sebelum restore untuk safety
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package restore

import (
	"context"
	"fmt"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/ui"
	"time"
)

// executePreBackup melakukan backup database sebelum restore sebagai safety net
// Menggunakan existing backup infrastructure (backup.Service)
// Returns: backup file path dan error
func (s *Service) executePreBackup(ctx context.Context, targetDB string) (string, error) {
	s.Log.Info("=== Pre-Restore Backup ===")
	s.Log.Infof("Membuat backup database '%s' sebelum restore...", targetDB)

	ui.PrintSubHeader("Pre-Restore Backup")
	fmt.Printf("Creating safety backup for: %s\n\n", targetDB)

	startTime := time.Now()

	// Build BackupDBOptions untuk backup service
	// Menggunakan struktur yang sama dengan backup command
	backupOpts := types.BackupDBOptions{
		Filter: types.FilterOptions{
			IncludeDatabases: []string{targetDB}, // Backup hanya database yang akan di-restore
			ExcludeDatabases: []string{},
		},
		Profile: types.ProfileInfo{
			Path:   s.TargetProfile.Path,
			Name:   s.TargetProfile.Name,
			DBInfo: s.TargetProfile.DBInfo,
		},
		DryRun: false,
		Force:  false,
		Entry: types.BackupEntryConfig{
			HeaderTitle: fmt.Sprintf("Pre-Restore Backup - %s", targetDB),
			Force:       false,
			SuccessMsg:  "",
			LogPrefix:   "[Pre-Restore Backup]",
			BackupMode:  "separated", // Gunakan separated mode untuk single database
		},
	}

	// Inisialisasi backup service menggunakan existing constructor
	backupSvc := backup.NewBackupService(s.Log, s.Config, &backupOpts)

	// Set cancel function untuk graceful shutdown (tanpa timeout)
	// Backup bisa memakan waktu lama tergantung ukuran database
	backupCancel := func() {}
	backupSvc.SetCancelFunc(backupCancel)

	// Execute backup menggunakan existing infrastructure
	// Gunakan ExecuteBackupWithResult yang akan kita tambahkan ke backup package
	// Atau kita call setupBackupConnections + ExecuteBackup secara manual

	// Setup backup session (internal call sequence sama dengan ExecuteBackupCommand)
	sourceClient, dbFiltered, err := backupSvc.PrepareBackupSession(ctx, "", false)
	if err != nil {
		return "", fmt.Errorf("gagal setup backup session: %w", err)
	}
	defer sourceClient.Close()

	// Execute backup dan dapatkan result
	result, err := backupSvc.ExecuteBackup(ctx, sourceClient, dbFiltered, "separated")
	if err != nil {
		return "", fmt.Errorf("pre-backup gagal: %w", err)
	}

	duration := time.Since(startTime)
	s.Log.Infof("✓ Pre-backup selesai dalam %s", duration)

	// Ambil backup file path dari result
	backupFilePath := ""
	if result != nil && len(result.BackupInfo) > 0 {
		// Ambil path file dari backup info pertama (single database backup)
		backupFilePath = result.BackupInfo[0].OutputFile
	}

	if backupFilePath == "" {
		s.Log.Warn("Backup berhasil tapi tidak dapat menentukan path file backup")
		return "", nil // Tidak error, tapi tidak ada rollback capability
	}

	ui.PrintSuccess(fmt.Sprintf("Safety backup created: %s", backupFilePath))
	fmt.Println()

	return backupFilePath, nil
}
