package backup

import (
	"fmt"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/profilehelper"
	"sfDBTools/pkg/ui"
)

// CheckAndSelectConfigFile memeriksa file konfigurasi yang ada atau memandu pengguna untuk memilihnya
func (s *Service) CheckAndSelectConfigFile() error {
	allowInteractive := isProfileSelectionInteractiveMode(s.BackupDBOptions.Mode) && s.BackupDBOptions.Profile.Path == ""
	profile, err := profilehelper.LoadSourceProfile(
		s.Config.ConfigDir.DatabaseProfile,
		s.BackupDBOptions.Profile.Path,
		s.BackupDBOptions.Encryption.Key,
		allowInteractive,
	)
	if err != nil {
		return fmt.Errorf("gagal load source profile: %w", err)
	}

	s.BackupDBOptions.Profile = *profile
	return nil
}

// SetupBackupExecution mempersiapkan konfigurasi backup yang umum
func (s *Service) SetupBackupExecution() error {
	ui.PrintSubHeader("Persiapan Eksekusi Backup")

	// Prompt untuk ticket jika tidak di-provide
	if s.BackupDBOptions.Ticket == "" {
		s.Log.Info("Ticket number tidak ditemukan, meminta input...")
		ticket, err := input.AskTicket(consts.FeatureBackup)
		if err != nil {
			return fmt.Errorf("gagal mendapatkan ticket number: %w", err)
		}
		s.BackupDBOptions.Ticket = ticket
		s.Log.Infof("Ticket number: %s", s.BackupDBOptions.Ticket)
	} else {
		s.Log.Infof("Ticket number: %s", s.BackupDBOptions.Ticket)
	}

	// Membuat direktori output jika belum ada
	s.Log.Info("Membuat direktori output jika belum ada : " + s.BackupDBOptions.OutputDir)
	if err := fsops.CreateDirIfNotExist(s.BackupDBOptions.OutputDir); err != nil {
		return fmt.Errorf("gagal membuat direktori output: %w", err)
	}
	s.Log.Info("Direktori output siap: " + s.BackupDBOptions.OutputDir)

	// Log konfigurasi
	if s.BackupDBOptions.Encryption.Enabled {
		s.Log.Info("Enkripsi AES-256-GCM diaktifkan untuk backup (kompatibel dengan OpenSSL)")
	} else {
		s.Log.Info("Enkripsi tidak diaktifkan, melewati langkah kunci enkripsi...")
	}

	if s.BackupDBOptions.Compression.Enabled {
		s.Log.Infof("Kompresi %s diaktifkan (level: %d)", s.BackupDBOptions.Compression.Type, s.BackupDBOptions.Compression.Level)
	} else {
		s.Log.Info("Kompresi tidak diaktifkan, melewati langkah kompresi...")
	}

	if s.BackupDBOptions.Filter.ExcludeData {
		s.Log.Info("Opsi exclude-data diaktifkan: hanya struktur database yang akan di-backup.")
	} else {
		s.Log.Info("Data database akan disertakan dalam backup.")
	}

	return nil
}
