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
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/fsops"
	pkghelper "sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/profilehelper"
	"sfDBTools/pkg/ui"
	"strings"
)

// CheckAndSelectConfigFile memeriksa file konfigurasi yang ada atau memandu pengguna untuk memilihnya
func (s *Service) CheckAndSelectConfigFile() error {
	allowInteractive := (s.BackupDBOptions.Mode == "single" || s.BackupDBOptions.Mode == "primary" || s.BackupDBOptions.Mode == "secondary" || s.BackupDBOptions.Mode == "combined" || s.BackupDBOptions.Mode == "separated") && s.BackupDBOptions.Profile.Path == ""
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

	// Untuk mode single/primary/secondary, update stats setelah filtering kandidat
	if isSingleModeVariant(s.BackupDBOptions.Mode) {
		// Filter candidates untuk mendapatkan jumlah yang akurat
		candidates := s.filterCandidatesByMode(dbFiltered, s.BackupDBOptions.Mode)
		stats.TotalIncluded = len(candidates)
		stats.TotalExcluded = stats.TotalFound - len(candidates)
	}

	display.DisplayFilterStats(stats, s.Log)

	// Generate output directory dan filename
	// Untuk mode single/primary/secondary, dbFiltered akan di-update dengan database yang dipilih + companion
	dbFiltered, err = s.generateBackupPaths(ctx, client, dbFiltered)
	if err != nil {
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
// Untuk command filter tanpa include/exclude flags: tampilkan multi-select
// Untuk command all atau filter dengan flags: gunakan filter biasa
func (s *Service) GetFilteredDatabases(ctx context.Context, client *database.Client) ([]string, *types.DatabaseFilterStats, error) {
	hasIncludeFlags := len(s.BackupDBOptions.Filter.IncludeDatabases) > 0 || s.BackupDBOptions.Filter.IncludeFile != ""

	// Jika ada include flags, selalu gunakan filter biasa (tidak perlu multi-select)
	if hasIncludeFlags {
		return database.FilterFromBackupOptions(ctx, client, s.BackupDBOptions)
	}

	// Jika ini command filter (IsFilterCommand=true) dan tidak ada include/exclude yang di-set manual
	// maka tampilkan multi-select
	isFilterMode := s.BackupDBOptions.Filter.IsFilterCommand
	hasAnyExcludeConfig := len(s.BackupDBOptions.Filter.ExcludeDatabases) > 0 ||
		s.BackupDBOptions.Filter.ExcludeDBFile != ""

	// Untuk command filter tanpa include dan exclude manual → multi-select
	if isFilterMode && !hasAnyExcludeConfig {
		return s.getFilteredDatabasesWithMultiSelect(ctx, client)
	}

	// Untuk command all atau filter dengan flags → gunakan filter biasa dengan nilai default
	return database.FilterFromBackupOptions(ctx, client, s.BackupDBOptions)
}

// getFilteredDatabasesWithMultiSelect menampilkan multi-select untuk memilih database
func (s *Service) getFilteredDatabasesWithMultiSelect(ctx context.Context, client *database.Client) ([]string, *types.DatabaseFilterStats, error) {
	// Get all databases from server
	allDatabases, err := client.GetDatabaseList(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mengambil daftar database: %w", err)
	}

	stats := &types.DatabaseFilterStats{
		TotalFound:    len(allDatabases),
		TotalIncluded: 0,
		TotalExcluded: 0,
	}

	if len(allDatabases) == 0 {
		return nil, stats, fmt.Errorf("tidak ada database yang ditemukan di server")
	}

	// Filter system databases untuk pilihan
	nonSystemDBs := make([]string, 0, len(allDatabases))
	systemDBs := []string{"information_schema", "performance_schema", "mysql", "sys"}

	for _, db := range allDatabases {
		isSystem := false
		for _, sysDB := range systemDBs {
			if strings.EqualFold(db, sysDB) {
				isSystem = true
				break
			}
		}
		if !isSystem {
			nonSystemDBs = append(nonSystemDBs, db)
		}
	}

	if len(nonSystemDBs) == 0 {
		return nil, stats, fmt.Errorf("tidak ada database non-system yang tersedia untuk dipilih")
	}

	// Tampilkan multi-select
	ui.PrintSubHeader("Pilih Database untuk Backup")
	selectedDBs, err := s.selectMultipleDatabases(nonSystemDBs)
	if err != nil {
		return nil, stats, err
	}

	if len(selectedDBs) == 0 {
		return nil, stats, fmt.Errorf("tidak ada database yang dipilih")
	}

	stats.TotalIncluded = len(selectedDBs)
	stats.TotalExcluded = len(allDatabases) - len(selectedDBs)

	return selectedDBs, stats, nil
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
// Returns updated dbFiltered untuk mode single/primary/secondary (database yang dipilih + companion)
func (s *Service) generateBackupPaths(ctx context.Context, client *database.Client, dbFiltered []string) ([]string, error) {
	dbHostname := s.BackupDBOptions.Profile.DBInfo.Host
	compressionSettings := s.getCompressionSettings()

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
	dbCount := 0
	if s.BackupDBOptions.Mode == "separated" || s.BackupDBOptions.Mode == "separate" ||
		isSingleModeVariant(s.BackupDBOptions.Mode) {
		exampleDBName = "database_name"
	} else if s.BackupDBOptions.Mode == "combined" {
		// Untuk combined, gunakan jumlah database yang akan di-backup
		dbCount = len(dbFiltered)
	}

	s.BackupDBOptions.File.Path, err = pkghelper.GenerateBackupFilenameWithCount(
		exampleDBName,
		s.BackupDBOptions.Mode,
		dbHostname,
		compressionSettings.Type,
		s.BackupDBOptions.Encryption.Enabled,
		dbCount,
	)
	if err != nil {
		s.Log.Warn("gagal generate filename preview: " + err.Error())
		s.BackupDBOptions.File.Path = "error_generating_filename"
	}

	// Handle single/primary/secondary mode dengan database selection
	if isSingleModeVariant(s.BackupDBOptions.Mode) {
		return s.handleSingleModeSetup(ctx, client, dbFiltered, compressionSettings)
	}

	return dbFiltered, nil
}

// handleSingleModeSetup handle setup untuk mode single/primary/secondary
// Returns companionDbs (database yang dipilih + companion) sebagai dbFiltered yang baru
func (s *Service) handleSingleModeSetup(ctx context.Context, client *database.Client, dbFiltered []string, compressionSettings types_backup.CompressionSettings) ([]string, error) {
	companionDbs, selectedDB, companionStatus, selErr := s.selectDatabaseAndBuildList(
		ctx, client, s.BackupDBOptions.DBName, dbFiltered, s.BackupDBOptions.Mode)
	if selErr != nil {
		return nil, selErr
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

	// Return companionDbs sebagai dbFiltered yang baru
	// companionDbs berisi: [database_yang_dipilih, companion_dmart, companion_temp, companion_archive]
	return companionDbs, nil
}

// selectMultipleDatabases menampilkan multi-select menu untuk memilih database
func (s *Service) selectMultipleDatabases(databases []string) ([]string, error) {
	if len(databases) == 0 {
		return nil, fmt.Errorf("tidak ada database yang tersedia untuk dipilih")
	}

	s.Log.Info(fmt.Sprintf("Tersedia %d database non-system", len(databases)))
	s.Log.Info("Gunakan [Space] untuk memilih/membatalkan, [Enter] untuk konfirmasi")

	// Gunakan ShowMultiSelect dari input package
	indices, err := input.ShowMultiSelect("Pilih database untuk backup:", databases)
	if err != nil {
		return nil, fmt.Errorf("pemilihan database dibatalkan: %w", err)
	}

	if len(indices) == 0 {
		return nil, fmt.Errorf("tidak ada database yang dipilih")
	}

	// Convert indices to database names
	selectedDBs := make([]string, 0, len(indices))
	for _, idx := range indices {
		if idx > 0 && idx <= len(databases) {
			selectedDBs = append(selectedDBs, databases[idx-1])
		}
	}

	s.Log.Info(fmt.Sprintf("Dipilih %d database: %s", len(selectedDBs), strings.Join(selectedDBs, ", ")))

	return selectedDBs, nil
}
