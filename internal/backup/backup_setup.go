package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/profileselect"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"
)

// setupBackupConnections melakukan setup koneksi source
func (s *Service) setupBackupConnections(ctx context.Context, headerTitle string, showOptions bool) (*database.Client, []string, func(), error) {

	// Setup session (koneksi database source) untuk mode normal
	sourceClient, dbFiltered, err := s.PrepareBackupSession(ctx, headerTitle, showOptions)
	if err != nil {
		return nil, nil, nil, err
	}

	// Cleanup function untuk close semua connections
	cleanup := func() {
		if sourceClient != nil {
			sourceClient.Close()
		}
	}

	return sourceClient, dbFiltered, cleanup, nil
}

// CheckAndSelectConfigFile memeriksa file konfigurasi yang ada atau memandu pengguna untuk memilihnya.
// Fungsi ini sekarang menggunakan fungsi generic dari pkg/dbconfig untuk menghindari duplikasi kode.
func (s *Service) CheckAndSelectConfigFile() error {
	// Jika user sudah memberikan path profile via flag/ENV, gunakan itu dan jangan prompt
	if strings.TrimSpace(s.BackupDBOptions.Profile.Path) != "" {
		absPath, name, err := helper.ResolveConfigPath(s.BackupDBOptions.Profile.Path)
		if err != nil {
			return fmt.Errorf("gagal memproses path konfigurasi: %w", err)
		}

		// Muat dan parse profil terenkripsi menggunakan key (jika ada)
		loaded, err := profileselect.LoadAndParseProfile(absPath, s.BackupDBOptions.Encryption.Key)
		if err != nil {
			return err
		}

		// Simpan kembali ke BackupDBOptions (pertahankan metadata yang relevan)
		s.BackupDBOptions.Profile.Path = absPath
		s.BackupDBOptions.Profile.Name = name
		s.BackupDBOptions.Profile.DBInfo = loaded.DBInfo
		s.BackupDBOptions.Profile.EncryptionSource = loaded.EncryptionSource
		return nil
	}

	// Jika tidak ada path diberikan, buka selector agar user memilih file profil
	info, err := profileselect.SelectExistingDBConfig("Pilih file konfigurasi database sumber:")
	if err != nil {
		return fmt.Errorf("gagal memilih konfigurasi database: %w", err)
	}
	// Simpan ke BackupDBOptions
	s.BackupDBOptions.Profile = info
	return nil
}

// SetupBackupExecution mempersiapkan konfigurasi backup yang umum
func (s *Service) SetupBackupExecution() error {
	ui.PrintSubHeader("Persiapan Eksekusi Backup")

	// Membuat direktori output jika belum ada
	s.Log.Info("Membuat direktori output jika belum ada : " + s.BackupDBOptions.OutputDir)
	created, err := fsops.CreateBackupDirs(s.BackupDBOptions.OutputDir, s.Config.Backup.Output.Structure.CreateSubdirs, s.Config.Backup.Output.Structure.Pattern, s.Config.General.ClientCode)
	if err != nil {
		return fmt.Errorf("gagal membuat direktori output: %w", err)
	}
	if created {
		s.Log.Info("Direktori output siap: " + s.BackupDBOptions.OutputDir)
	} else {
		s.Log.Info("Direktori output sudah ada: " + s.BackupDBOptions.OutputDir)
	}

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
