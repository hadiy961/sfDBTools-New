// File : internal/backup/backup_entry.go
// Deskripsi : Entry points untuk semua jenis backup database
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-15
// Last Modified : 2024-10-15

package backup

import (
	"context"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/ui"
)

// ExecuteBackupCommand adalah unified entry point untuk semua jenis backup
func (s *Service) ExecuteBackupCommand(ctx context.Context, config types.BackupEntryConfig) error {
	// Setup session (koneksi database source)
	sourceClient, dbFiltered, err := s.PrepareBackupSession(ctx, config.HeaderTitle, config.Force)
	if err != nil {
		return err
	}

	// Cleanup function untuk close semua connections
	cleanup := func() {
		if sourceClient != nil {
			sourceClient.Close()
		}
	}

	defer cleanup()

	// Lakukan backup
	result, err := s.ExecuteBackup(ctx, sourceClient, dbFiltered, config.BackupMode)
	if err != nil {
		// s.Log.Error(config.LogPrefix + " gagal: " + err.Error())
		return err
	}

	// Tampilkan hasil
	s.DisplayBackupResult(result)

	// Print success message jika ada
	if config.SuccessMsg != "" {
		ui.PrintSuccess(config.SuccessMsg)
		s.Log.Info(config.SuccessMsg)
	}
	return nil
}

// BackupSeparate melakukan backup database dengan file terpisah per database
func (s *Service) BackupSeparate(ctx context.Context) error {
	config := types.BackupEntryConfig{
		HeaderTitle: "Backup Database (Hasil Backup Database Terpisah)",
		Force:       s.BackupDBOptions.Force,
		BackupMode:  "separate",
		SuccessMsg:  "",
		LogPrefix:   "[Separate Backup] ",
	}
	return s.ExecuteBackupCommand(ctx, config)
}

// BackupCombined melakukan backup semua database dalam satu file
func (s *Service) BackupCombined(ctx context.Context) error {
	config := types.BackupEntryConfig{
		HeaderTitle: "Backup Database (Hasil Backup Database Digabung)",
		Force:       s.BackupDBOptions.Force,
		BackupMode:  "combined",
		SuccessMsg:  "backup semua database selesai.",
		LogPrefix:   "[Combined Backup] ",
	}
	return s.ExecuteBackupCommand(ctx, config)
}
