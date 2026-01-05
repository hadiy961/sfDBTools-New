package restore

// File : internal/restore/setup_shared_backup.go
// Deskripsi : Helper konfigurasi backup pre-restore
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

import "sfDBTools/internal/types"

func (s *Service) setupBackupOptions(backupOpts *types.RestoreBackupOptions, encryptionKey string, allowInteractive bool) {
	if backupOpts.OutputDir == "" {
		backupOpts.OutputDir = s.getBackupDirectory(allowInteractive)
	}

	if !backupOpts.Compression.Enabled {
		backupOpts.Compression = types.CompressionOptions{
			Enabled: s.Config.Backup.Compression.Enabled,
			Type:    s.Config.Backup.Compression.Type,
			Level:   s.Config.Backup.Compression.Level,
		}
	}

	if !backupOpts.Encryption.Enabled {
		backupOpts.Encryption = types.EncryptionOptions{
			Enabled: s.Config.Backup.Encryption.Enabled,
			Key:     encryptionKey,
		}
	}
}
