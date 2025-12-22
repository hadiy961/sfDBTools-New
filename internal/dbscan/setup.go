// File : internal/dbscan/setup.go
// Deskripsi : Setup connection, configuration loading, dan session preparation
// Author : Hadiyatna Muflihun
// Tanggal : 17 Desember 2025
// Last Modified : 17 Desember 2025

package dbscan

import (
	"context"
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/profilehelper"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

// ResolveScanLists membaca file include/exclude dari path yang ada di ScanOptions
// dan menggabungkannya ke dalam IncludeList/ExcludeList.
func ResolveScanLists(opts *types.ScanOptions) error {
	// Include File
	if opts.DatabaseList.File != "" {
		lines, err := fsops.ReadLinesFromFile(opts.DatabaseList.File)
		if err != nil {
			return fmt.Errorf("gagal membaca db-file %s: %w", opts.DatabaseList.File, err)
		}
		opts.IncludeList = append(opts.IncludeList, helper.ListTrimNonEmpty(lines)...)
	}

	// Exclude File
	if opts.ExcludeFile != "" {
		lines, err := fsops.ReadLinesFromFile(opts.ExcludeFile)
		if err != nil {
			return fmt.Errorf("gagal membaca exclude-file %s: %w", opts.ExcludeFile, err)
		}
		opts.ExcludeList = append(opts.ExcludeList, helper.ListTrimNonEmpty(lines)...)
	}

	// Deduplicate lists
	opts.IncludeList = helper.ListUnique(opts.IncludeList)
	opts.ExcludeList = helper.ListUnique(opts.ExcludeList)

	// Validation: minimal ada kriteria filter
	if len(opts.IncludeList) == 0 && len(opts.ExcludeList) == 0 {
		return fmt.Errorf("minimal salah satu flag harus digunakan: gunakan --db/--db-file untuk include atau --exclude-db/--exclude-file untuk exclude")
	}

	// Logic: Include - Exclude
	// Jika user memberikan include list DAN exclude list, maka kita kurangi include list dengan exclude list.
	// Jika hanya exclude list, maka include list kosong (artinya scan semua kecuali exclude, ini ditangani di layer database filtering).
	if len(opts.IncludeList) > 0 && len(opts.ExcludeList) > 0 {
		opts.IncludeList = helper.ListSubtract(opts.IncludeList, opts.ExcludeList)
		opts.ExcludeList = []string{} // Reset exclude list karena sudah diaplikasikan ke include list
	}

	return nil
}

// setupScanConnections mengorkestrasi setup koneksi source dan target database.
// Menangani mode normal dan mode rescan.
// Returns: sourceClient, targetClient, dbFiltered, cleanupFunc, error
func (s *Service) setupScanConnections(ctx context.Context, headerTitle string, showOptions bool) (*database.Client, *database.Client, []string, func(), error) {
	// Mode Rescan: penanganan khusus
	if s.ScanOptions.Mode == "rescan" {
		return s.prepareRescanSession(ctx, headerTitle, showOptions)
	}

	// Mode Normal: setup standar
	sourceClient, dbFiltered, err := s.prepareScanSession(ctx, headerTitle, showOptions)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Setup Target Database Connection (Optional based on config)
	var targetClient *database.Client
	if s.ScanOptions.SaveToDB {
		targetClient, err = s.ConnectToTargetDB(ctx)
		if err != nil {
			s.Log.Warn("Gagal koneksi ke target database, hasil scan tidak akan disimpan: " + err.Error())
			s.ScanOptions.SaveToDB = false
		}
	}

	// Cleanup function untuk menutup semua koneksi
	cleanup := func() {
		if sourceClient != nil {
			sourceClient.Close()
		}
		if targetClient != nil {
			targetClient.Close()
		}
	}

	return sourceClient, targetClient, dbFiltered, cleanup, nil
}

// prepareScanSession mempersiapkan session untuk scanning normal.
// Meload config, connect ke source, dan filter database.
func (s *Service) prepareScanSession(ctx context.Context, headerTitle string, showOptions bool) (*database.Client, []string, error) {
	if headerTitle != "" {
		ui.Headers(headerTitle)
		s.Log.Infof("=== %s ===", headerTitle)
	}

	if err := s.CheckAndSelectConfigFile(); err != nil {
		return nil, nil, err
	}

	// Validasi Local Scan
	if s.ScanOptions.LocalScan && s.ScanOptions.ProfileInfo.DBInfo.Host != "localhost" {
		return nil, nil, fmt.Errorf("local scan hanya didukung untuk host 'localhost'")
	}

	// Connect ke Source Database
	client, err := profilehelper.ConnectWithProfile(&s.ScanOptions.ProfileInfo, consts.DefaultInitialDatabase)
	if err != nil {
		return nil, nil, err
	}

	// Defer close jika terjadi error dalam fungsi ini
	var success bool
	defer func() {
		if !success && client != nil {
			client.Close()
		}
	}()

	// Get & Filter Databases
	dbFiltered, stats, err := s.GetFilteredDatabases(ctx, client)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mendapatkan daftar database: %w", err)
	}

	// Show Options / Stats if requested
	if showOptions {
		s.DisplayFilterStats(stats)
		if proceed, askErr := s.DisplayScanOptions(); askErr != nil {
			return nil, nil, askErr
		} else if !proceed {
			return nil, nil, validation.ErrUserCancelled
		}
	}

	if len(dbFiltered) == 0 {
		return nil, nil, fmt.Errorf("tidak ada database untuk di-scan setelah filtering")
	}

	success = true
	return client, dbFiltered, nil
}

// prepareRescanSession mempersiapkan session untuk mode rescan.
// Mengambil daftar database yang gagal dari target database.
func (s *Service) prepareRescanSession(ctx context.Context, headerTitle string, showOptions bool) (*database.Client, *database.Client, []string, func(), error) {
	if headerTitle != "" {
		ui.Headers(headerTitle)
		s.Log.Infof("=== %s ===", headerTitle)
	}

	if showOptions {
		if proceed, askErr := s.DisplayScanOptions(); askErr != nil {
			return nil, nil, nil, nil, askErr
		} else if !proceed {
			return nil, nil, nil, nil, validation.ErrUserCancelled
		}
	}

	if err := s.CheckAndSelectConfigFile(); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("gagal memuat konfigurasi database: %w", err)
	}

	// Connect ke Target Database (Wajib untuk rescan)
	targetClient, err := s.ConnectToTargetDB(ctx)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("gagal koneksi ke target database: %w", err)
	}

	// Helper cleanup sementara
	var success bool
	defer func() {
		if !success && targetClient != nil {
			targetClient.Close()
		}
	}()

	// Ambil list failed databases
	failedDBNames, err := database.GetFailedDatabaseNames(ctx, targetClient)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("gagal mendapatkan list database yang gagal: %w", err)
	}

	if len(failedDBNames) == 0 {
		ui.PrintInfo("Tidak ada database yang gagal untuk di-rescan")
		return nil, nil, nil, nil, fmt.Errorf("tidak ada database yang gagal untuk di-rescan")
	}

	ui.PrintInfo(fmt.Sprintf("Ditemukan %d database yang gagal di-scan sebelumnya", len(failedDBNames)))

	// Connect ke Source Database
	sourceClient, err := profilehelper.ConnectWithProfile(&s.ScanOptions.ProfileInfo, consts.DefaultInitialDatabase)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("gagal koneksi ke source database: %w", err)
	}

	// Display simplified stats
	stats := &types.FilterStats{
		TotalFound:    len(failedDBNames),
		TotalIncluded: len(failedDBNames),
	}
	s.DisplayFilterStats(stats)

	// Force enable SaveToDB untuk rescan (update status error)
	s.ScanOptions.SaveToDB = true

	cleanup := func() {
		if sourceClient != nil {
			sourceClient.Close()
		}
		if targetClient != nil {
			targetClient.Close()
		}
	}

	success = true
	return sourceClient, targetClient, failedDBNames, cleanup, nil
}

// CheckAndSelectConfigFile memeriksa atau memilih file profile database.
func (s *Service) CheckAndSelectConfigFile() error {
	profile, err := profilehelper.LoadSourceProfile(
		s.Config.ConfigDir.DatabaseProfile,
		s.ScanOptions.ProfileInfo.Path,
		s.ScanOptions.Encryption.Key,
		true, // enableInteractive
	)
	if err != nil {
		return fmt.Errorf("gagal load source profile: %w", err)
	}

	s.ScanOptions.ProfileInfo = *profile
	return nil
}

// ConnectToTargetDB membuat koneksi ke database pusat (app database).
func (s *Service) ConnectToTargetDB(ctx context.Context) (*database.Client, error) {
	client, err := database.ConnectToAppDatabase()
	if err != nil {
		return nil, fmt.Errorf("gagal koneksi ke target database: %w", err)
	}

	if err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("gagal verifikasi koneksi target: %w", err)
	}

	return client, nil
}

// getTargetDBConfig mengambil konfigurasi target database (untuk display options).
func (s *Service) getTargetDBConfig() types.ServerDBConnection {
	conn := s.ScanOptions.TargetDB

	// Fallback ke env defaults jika kosong
	if conn.Host == "" {
		conn.Host = helper.GetEnvOrDefault("SFDB_DB_HOST", "localhost")
	}
	if conn.Port == 0 {
		conn.Port = helper.GetEnvOrDefaultInt("SFDB_DB_PORT", 3306)
	}
	if conn.User == "" {
		conn.User = helper.GetEnvOrDefault("SFDB_DB_USER", "root")
	}
	if conn.Database == "" {
		conn.Database = helper.GetEnvOrDefault("SFDB_DB_NAME", "sfDBTools")
	}
	// Password tidak perlu default, biarkan kosong jika tidak diset

	return types.ServerDBConnection{
		Host:     conn.Host,
		Port:     conn.Port,
		User:     conn.User,
		Password: conn.Password,
		Database: conn.Database,
	}
}

// GetFilteredDatabases mengambil dan memfilter daftar database.
func (s *Service) GetFilteredDatabases(ctx context.Context, client *database.Client) ([]string, *types.FilterStats, error) {
	return database.FilterFromScanOptions(ctx, client, &s.ScanOptions)
}
