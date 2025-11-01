package dbscan

import (
	"context"
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
)

// PrepareScanSession mengatur seluruh alur persiapan sebelum proses scanning dimulai.
// Fungsi ini sekarang lebih tangguh dalam menangani resource (koneksi database)
// dengan menggunakan `defer` untuk memastikan koneksi ditutup jika terjadi kegagalan.
func (s *Service) PrepareScanSession(ctx context.Context, headerTitle string, showOptions bool) (client *database.Client, dbFiltered []string, err error) {
	if headerTitle != "" {
		ui.Headers(headerTitle)
		s.Logger.Infof("=== %s ===", headerTitle)
	}

	if err = s.CheckAndSelectConfigFile(); err != nil {
		return nil, nil, err
	}

	// Gunakan helper ConnectToSourceDatabase agar konsisten dan teruji
	creds := types.SourceDBConnection{
		DBInfo:   s.ScanOptions.ProfileInfo.DBInfo,
		Database: "mysql", // gunakan schema sistem untuk koneksi awal
	}

	if s.ScanOptions.LocalScan && creds.DBInfo.Host != "localhost" {
		return nil, nil, fmt.Errorf("local scan hanya didukung untuk host 'localhost'")
	}

	client, err = database.ConnectToSourceDatabase(creds)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal koneksi ke database: %w", err)
	}

	// Gunakan pola `defer` dengan flag untuk memastikan `client.Close()` hanya dipanggil saat terjadi error.
	// Jika fungsi berhasil, client akan dikembalikan dalam keadaan terbuka.
	var success bool
	defer func() {
		if !success && client != nil {
			client.Close()
		}
	}()

	var stats *types.DatabaseFilterStats
	dbFiltered, stats, err = s.GetFilteredDatabases(ctx, client)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mendapatkan daftar database: %w", err)
	}

	if showOptions {
		s.DisplayFilterStats(stats)
		if proceed, askErr := s.DisplayScanOptions(); askErr != nil {
			return nil, nil, askErr
		} else if !proceed {
			return nil, nil, types.ErrUserCancelled
		}
	}

	if len(dbFiltered) == 0 {
		return nil, nil, fmt.Errorf("tidak ada database untuk di-scan setelah filtering")
	}

	success = true // Tandai sebagai sukses agar koneksi tidak ditutup oleh defer.
	return client, dbFiltered, nil
}
