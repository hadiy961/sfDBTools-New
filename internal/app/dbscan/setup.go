// File : internal/app/dbscan/setup.go
// Deskripsi : Setup connection, configuration loading, dan session preparation
// Author : Hadiyatna Muflihun
// Tanggal : 17 Desember 2025
// Last Modified : 05 Januari 2026

package dbscan

import (
	"context"
	"fmt"
	"sfDBTools/internal/app/dbscan/helpers"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	profilehelper "sfDBTools/pkg/helper/profile"
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

// setupScanConnections mengorkestrasi setup koneksi source database.
// Catatan: sfDBTools tidak menyimpan hasil scan ke database, jadi tidak ada koneksi target.
// Returns: sourceClient, dbFiltered, cleanupFunc, error
func (s *Service) setupScanConnections(ctx context.Context, headerTitle string, showOptions bool) (*database.Client, []string, func(), error) {
	// Mode Normal: setup standar
	sourceClient, dbFiltered, err := s.prepareScanSession(ctx, headerTitle, showOptions)
	if err != nil {
		return nil, nil, nil, err
	}

	// Cleanup function untuk menutup semua koneksi
	cleanup := func() {
		if sourceClient != nil {
			sourceClient.Close()
		}
	}

	return sourceClient, dbFiltered, cleanup, nil
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

// GetFilteredDatabases mengambil dan memfilter daftar database.
func (s *Service) GetFilteredDatabases(ctx context.Context, client *database.Client) ([]string, *types.FilterStats, error) {
	return helpers.FilterFromScanOptions(ctx, client, &s.ScanOptions)
}
