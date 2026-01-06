package restore

import (
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/domain"
)

// File : internal/restore/setup_shared_backup.go
// Deskripsi : Helper konfigurasi backup pre-restore
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 5 Januari 2026
func (s *Service) setupBackupOptions(backupOpts *restoremodel.RestoreBackupOptions, encryptionKey string, allowInteractive bool) {
	if backupOpts.OutputDir == "" {
		backupOpts.OutputDir = s.getBackupDirectory(allowInteractive)
	}

	if !backupOpts.Compression.Enabled {
		backupOpts.Compression = domain.CompressionOptions{
			Enabled: s.Config.Backup.Compression.Enabled,
			Type:    s.Config.Backup.Compression.Type,
			Level:   s.Config.Backup.Compression.Level,
		}
	}

	if !backupOpts.Encryption.Enabled {
		backupOpts.Encryption = domain.EncryptionOptions{
			Enabled: s.Config.Backup.Encryption.Enabled,
			Key:     encryptionKey,
		}
	}
}
