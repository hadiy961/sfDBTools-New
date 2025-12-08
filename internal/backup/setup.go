// File : internal/backup/setup.go
// Deskripsi : Setup dan preparation functions untuk backup operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/backup/display"
	"sfDBTools/internal/backup/helper"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/fsops"
	pkghelper "sfDBTools/pkg/helper"
	"sfDBTools/pkg/profilehelper"
	"sfDBTools/pkg/ui"
)

// CheckAndSelectConfigFile memeriksa file konfigurasi yang ada atau memandu pengguna untuk memilihnya
func (s *Service) CheckAndSelectConfigFile() error {
	allowInteractive := (s.BackupDBOptions.Mode == "single" || s.BackupDBOptions.Mode == "primary" || s.BackupDBOptions.Mode == "secondary") && s.BackupDBOptions.Profile.Path == ""
	profile, err := profilehelper.LoadSourceProfile(
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

// PrepareBackupSession mengatur seluruh alur persiapan sebelum proses backup dimulai
func (s *Service) PrepareBackupSession(ctx context.Context, headerTitle string, showOptions bool) (client *database.Client, dbFiltered []string, err error) {
	if headerTitle != "" {
		ui.Headers(headerTitle)
	}

	if err = s.CheckAndSelectConfigFile(); err != nil {
		return nil, nil, err
	}

	// Gunakan profilehelper untuk koneksi yang konsisten
	client, err = profilehelper.ConnectWithProfile(&s.BackupDBOptions.Profile, "mysql")
	if err != nil {
		return nil, nil, err
	}

	// Defer cleanup jika gagal
	var success bool
	defer func() {
		if !success && client != nil {
			client.Close()
		}
	}()

	// Ambil hostname dari MySQL server
	serverHostname, err := client.GetServerHostname(ctx)
	if err != nil {
		s.Log.Warnf("gagal mendapatkan hostname dari server: %v, menggunakan dari config", err)
		serverHostname = s.BackupDBOptions.Profile.DBInfo.Host
	} else {
		s.BackupDBOptions.Profile.DBInfo.HostName = serverHostname
		s.Log.Infof("menggunakan hostname dari server: %s", serverHostname)
	}

	var stats *types.DatabaseFilterStats
	dbFiltered, stats, err = s.GetFilteredDatabases(ctx, client)
	if err != nil {
		if stats != nil {
			display.DisplayFilterStats(stats, s.Log)
		}
		return nil, nil, fmt.Errorf("gagal mendapatkan daftar database: %w", err)
	}

	if len(dbFiltered) == 0 {
		display.DisplayFilterStats(stats, s.Log)
		ui.PrintError("Tidak ada database yang tersedia setelah filtering!")
		s.displayFilterWarnings(stats)
		return nil, nil, fmt.Errorf("tidak ada database tersedia untuk backup setelah filtering")
	}

	display.DisplayFilterStats(stats, s.Log)

	// Generate output directory dan filename
	if err := s.generateBackupPaths(ctx, client, dbFiltered); err != nil {
		return nil, nil, err
	}

	if !showOptions {
		if proceed, askErr := display.NewOptionsDisplayer(s.BackupDBOptions).Display(); askErr != nil {
			return nil, nil, askErr
		} else if !proceed {
			return nil, nil, types.ErrUserCancelled
		}
	}

	success = true
	return client, dbFiltered, nil
}

// GetFilteredDatabases mengambil dan memfilter daftar database sesuai aturan
func (s *Service) GetFilteredDatabases(ctx context.Context, client *database.Client) ([]string, *types.DatabaseFilterStats, error) {
	return database.FilterFromBackupOptions(ctx, client, s.BackupDBOptions)
}

// displayFilterWarnings menampilkan warning messages untuk filter stats
func (s *Service) displayFilterWarnings(stats *types.DatabaseFilterStats) {
	ui.PrintWarning("Kemungkinan penyebab:")

	if stats.TotalExcluded == stats.TotalFound {
		ui.PrintWarning(fmt.Sprintf("  • Semua database (%d) dikecualikan oleh filter exclude", stats.TotalExcluded))
	}

	if len(stats.NotFoundInInclude) > 0 {
		ui.PrintWarning("  • Database yang diminta di include list tidak ditemukan:")
		for _, db := range stats.NotFoundInInclude {
			ui.PrintWarning(fmt.Sprintf("    - %s", db))
		}
	}

	if len(stats.NotFoundInWhitelist) > 0 {
		ui.PrintWarning("  • Database dari whitelist file tidak ditemukan:")
		for _, db := range stats.NotFoundInWhitelist {
			ui.PrintWarning(fmt.Sprintf("    - %s", db))
		}
	}

	if len(stats.NotFoundInExclude) > 0 {
		ui.PrintWarning("  • Database yang diminta di exclude list tidak ditemukan:")
		for _, db := range stats.NotFoundInExclude {
			ui.PrintWarning(fmt.Sprintf("    - %s", db))
		}
	}

	if len(stats.NotFoundInBlacklist) > 0 {
		ui.PrintWarning("  • Database dari exclude file tidak ditemukan:")
		for _, db := range stats.NotFoundInBlacklist {
			ui.PrintWarning(fmt.Sprintf("    - %s", db))
		}
	}
}

// generateBackupPaths generate output directory dan filename untuk backup
func (s *Service) generateBackupPaths(ctx context.Context, client *database.Client, dbFiltered []string) error {
	dbHostname := s.BackupDBOptions.Profile.DBInfo.Host

	compressionSettings := helper.NewCompressionSettings(
		s.BackupDBOptions.Compression.Enabled,
		s.BackupDBOptions.Compression.Type,
		s.BackupDBOptions.Compression.Level,
	)

	// Generate output directory
	var err error
	s.BackupDBOptions.OutputDir, err = pkghelper.GenerateBackupDirectory(
		s.Config.Backup.Output.BaseDirectory,
		s.Config.Backup.Output.Structure.Pattern,
		dbHostname,
	)
	if err != nil {
		s.Log.Warn("gagal generate output directory: " + err.Error())
	}

	// Generate filename berdasarkan mode
	exampleDBName := ""
	if s.BackupDBOptions.Mode == "separated" || s.BackupDBOptions.Mode == "separate" ||
		s.BackupDBOptions.Mode == "single" || s.BackupDBOptions.Mode == "primary" ||
		s.BackupDBOptions.Mode == "secondary" {
		exampleDBName = "database_name"
	}

	s.BackupDBOptions.File.Path, err = pkghelper.GenerateBackupFilename(
		exampleDBName,
		s.BackupDBOptions.Mode,
		dbHostname,
		compressionSettings.Type,
		s.BackupDBOptions.Encryption.Enabled,
	)
	if err != nil {
		s.Log.Warn("gagal generate filename preview: " + err.Error())
		s.BackupDBOptions.File.Path = "error_generating_filename"
	}

	// Handle single/primary/secondary mode dengan database selection
	if s.BackupDBOptions.Mode == "single" || s.BackupDBOptions.Mode == "primary" || s.BackupDBOptions.Mode == "secondary" {
		return s.handleSingleModeSetup(ctx, client, dbFiltered, compressionSettings)
	}

	return nil
}

// handleSingleModeSetup handle setup untuk mode single/primary/secondary
func (s *Service) handleSingleModeSetup(ctx context.Context, client *database.Client, dbFiltered []string, compressionSettings types_backup.CompressionSettings) error {
	companionDbs, selectedDB, companionStatus, selErr := s.selectDatabaseAndBuildList(
		ctx, client, s.BackupDBOptions.DBName, dbFiltered, s.BackupDBOptions.Mode)
	if selErr != nil {
		return selErr
	}

	s.BackupDBOptions.DBName = selectedDB
	s.BackupDBOptions.CompanionStatus = companionStatus

	// Update filename untuk database yang dipilih
	previewFilename, err := pkghelper.GenerateBackupFilename(
		selectedDB,
		s.BackupDBOptions.Mode,
		s.BackupDBOptions.Profile.DBInfo.HostName,
		compressionSettings.Type,
		s.BackupDBOptions.Encryption.Enabled,
	)
	if err != nil {
		s.Log.Warn("gagal generate filename: " + err.Error())
		previewFilename = "error_generating_filename"
	}
	s.BackupDBOptions.File.Path = previewFilename

	// Update dbFiltered dengan companion databases
	copy(dbFiltered, companionDbs)

	return nil
}
