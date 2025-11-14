package backup

import (
	"fmt"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/profilehelper"
	"sfDBTools/pkg/ui"
)

// CheckAndSelectConfigFile memeriksa file konfigurasi yang ada atau memandu pengguna untuk memilihnya.
func (s *Service) CheckAndSelectConfigFile() error {
	// Load profile langsung tanpa interactive mode (profile harus sudah diisi di flag)
	profile, err := profilehelper.LoadSourceProfile(
		s.BackupDBOptions.Profile.Path,
		s.BackupDBOptions.Encryption.Key,
		false, // enableInteractive - tidak ada interactive selection, profile wajib diisi
	)
	if err != nil {
		return fmt.Errorf("gagal load source profile: %w", err)
	}

	// Simpan ke BackupDBOptions (dereference pointer)
	s.BackupDBOptions.Profile = *profile
	return nil
}

// SetupBackupExecution mempersiapkan konfigurasi backup yang umum
func (s *Service) SetupBackupExecution() error {
	ui.PrintSubHeader("Persiapan Eksekusi Backup")

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
