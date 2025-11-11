package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
)

// PrepareBackupSession mengatur seluruh alur persiapan sebelum proses backup dimulai.
// dengan menggunakan `defer` untuk memastikan koneksi ditutup jika terjadi kegagalan.
func (s *Service) PrepareBackupSession(ctx context.Context, headerTitle string, showOptions bool) (client *database.Client, dbFiltered []string, err error) {
	if headerTitle != "" {
		ui.Headers(headerTitle)
	}

	if err = s.CheckAndSelectConfigFile(); err != nil {
		return nil, nil, err
	}

	// Gunakan helper ConnectToSourceDatabase agar konsisten dan teruji
	creds := types.SourceDBConnection{
		DBInfo:   s.BackupDBOptions.Profile.DBInfo,
		Database: "mysql", // gunakan schema sistem untuk koneksi awal
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

	// Ambil hostname dari MySQL server (bukan dari config)
	serverHostname, err := client.GetServerHostname(ctx)
	if err != nil {
		s.Log.Warnf("gagal mendapatkan hostname dari server: %v, menggunakan dari config", err)
		serverHostname = s.BackupDBOptions.Profile.DBInfo.Host
	} else {
		// Update profile dengan hostname asli dari server
		s.BackupDBOptions.Profile.DBInfo.HostName = serverHostname
		s.Log.Infof("menggunakan hostname dari server: %s", serverHostname)
	}

	var stats *types.DatabaseFilterStats
	dbFiltered, stats, err = s.GetFilteredDatabases(ctx, client)
	if err != nil {
		// Tampilkan stats terlebih dahulu untuk memberikan konteks
		if stats != nil {
			s.DisplayFilterStats(stats)
		}
		return nil, nil, fmt.Errorf("gagal mendapatkan daftar database: %w", err)
	}

	if len(dbFiltered) == 0 {
		// Tampilkan stats sebelum error untuk memberikan informasi lengkap
		s.DisplayFilterStats(stats)

		ui.PrintError("Tidak ada database yang tersedia setelah filtering!")
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

		return nil, nil, fmt.Errorf("tidak ada database tersedia untuk backup setelah filtering")
	}

	s.DisplayFilterStats(stats)

	// Dapatkan hostname dari database server (bukan hostname lokal)
	dbHostname := s.BackupDBOptions.Profile.DBInfo.Host

	// Convert compression type dari string ke compress.CompressionType
	// Jika compression disabled, gunakan CompressionNone
	compressionType := compress.CompressionNone
	if s.BackupDBOptions.Compression.Enabled {
		compressionType = compress.CompressionType(s.BackupDBOptions.Compression.Type)
	}

	// Generate filename berdasarkan pattern dari config
	// Untuk mode separated, gunakan nama database contoh "database_name" agar lebih jelas
	exampleDBName := ""
	if s.BackupDBOptions.Mode == "separated" || s.BackupDBOptions.Mode == "separate" {
		exampleDBName = "database_name"
	}

	// Generate output directory berdasarkan config
	s.BackupDBOptions.OutputDir, err = helper.GenerateBackupDirectory(
		s.Config.Backup.Output.BaseDirectory,
		s.Config.Backup.Output.Structure.Pattern,
		dbHostname,
	)

	if err != nil {
		s.Log.Warn("gagal generate output directory: " + err.Error())
	}

	s.BackupDBOptions.File.Path, err = helper.GenerateBackupFilename(
		exampleDBName, // untuk preview: kosong untuk combined, "database_name" untuk separated
		s.BackupDBOptions.Mode,
		dbHostname,
		compressionType,
		s.BackupDBOptions.Encryption.Enabled,
	)

	if err != nil {
		s.Log.Warn("gagal generate filename preview: " + err.Error())
		s.BackupDBOptions.File.Path = "error_generating_filename"
	}

	if !showOptions {
		if proceed, askErr := s.DisplayBackupDBOptions(); askErr != nil {
			return nil, nil, askErr
		} else if !proceed {
			return nil, nil, types.ErrUserCancelled
		}
	}

	success = true // Tandai sebagai sukses agar koneksi tidak ditutup oleh defer.
	return client, dbFiltered, nil
}
