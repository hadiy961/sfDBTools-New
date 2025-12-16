// File : internal/dbscan/dbscan_prepare.go
// Deskripsi : Persiapan session untuk database scanning
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 16 Desember 2025

package dbscan

import (
	"context"
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/profilehelper"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

// PrepareScanSession mengatur seluruh alur persiapan sebelum proses scanning dimulai.
// Fungsi ini sekarang lebih tangguh dalam menangani resource (koneksi database)
// dengan menggunakan `defer` untuk memastikan koneksi ditutup jika terjadi kegagalan.
func (s *Service) PrepareScanSession(ctx context.Context, headerTitle string, showOptions bool) (client *database.Client, dbFiltered []string, err error) {
	if headerTitle != "" {
		ui.Headers(headerTitle)
		s.Log.Infof("=== %s ===", headerTitle)
	}

	if err = s.CheckAndSelectConfigFile(); err != nil {
		return nil, nil, err
	}

	// Check local scan restriction
	if s.ScanOptions.LocalScan && s.ScanOptions.ProfileInfo.DBInfo.Host != "localhost" {
		return nil, nil, fmt.Errorf("local scan hanya didukung untuk host 'localhost'")
	}

	// Gunakan profilehelper untuk koneksi yang konsisten
	client, err = profilehelper.ConnectWithProfile(&s.ScanOptions.ProfileInfo, "mysql")
	if err != nil {
		return nil, nil, err
	}

	// Gunakan pola `defer` dengan flag untuk memastikan `client.Close()` hanya dipanggil saat terjadi error.
	// Jika fungsi berhasil, client akan dikembalikan dalam keadaan terbuka.
	var success bool
	defer func() {
		if !success && client != nil {
			client.Close()
		}
	}()

	var stats *types.FilterStats
	dbFiltered, stats, err = s.GetFilteredDatabases(ctx, client)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mendapatkan daftar database: %w", err)
	}

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

	success = true // Tandai sebagai sukses agar koneksi tidak ditutup oleh defer.
	return client, dbFiltered, nil
}
