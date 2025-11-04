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
func (s *Service) ExecuteBackupCommand(config types.BackupEntryConfig) error {
	ctx := context.Background()

	// Setup connections
	sourceClient, dbFiltered, cleanup, err := s.setupBackupConnections(ctx, config.HeaderTitle, config.ShowOptions)
	if err != nil {
		return err
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
	}
	return nil
}

// BackupSeparate melakukan backup database dengan file terpisah per database
func (s *Service) BackupSeparate() error {
	config := types.BackupEntryConfig{
		HeaderTitle: "Backup Database (Hasil Backup Database Terpisah)",
		ShowOptions: s.BackupDBOptions.ShowOptions,
		BackupMode:  "separate",
		SuccessMsg:  "",
		LogPrefix:   "[Separate Backup] ",
	}
	return s.ExecuteBackupCommand(config)
}

// BackupCombined melakukan backup semua database dalam satu file
func (s *Service) BackupCombined() error {
	config := types.BackupEntryConfig{
		HeaderTitle: "Backup Database (Hasil Backup Database Digabung)",
		ShowOptions: s.BackupDBOptions.ShowOptions,
		BackupMode:  "combined",
		SuccessMsg:  "backup semua database selesai.",
		LogPrefix:   "[Combined Backup] ",
	}
	return s.ExecuteBackupCommand(config)
}
