package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/backup/display"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/profilehelper"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

// PrepareBackupSession mengatur seluruh alur persiapan sebelum proses backup dimulai
func (s *Service) PrepareBackupSession(ctx context.Context, headerTitle string, showOptions bool) (client *database.Client, dbFiltered []string, err error) {
	if headerTitle != "" {
		ui.Headers(headerTitle)
	}

	if err = s.CheckAndSelectConfigFile(); err != nil {
		return nil, nil, err
	}

	// Gunakan profilehelper untuk koneksi yang konsisten
	client, err = profilehelper.ConnectWithProfile(&s.BackupDBOptions.Profile, consts.DefaultInitialDatabase)
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

	var stats *types.FilterStats
	dbFiltered, stats, err = s.GetFilteredDatabases(ctx, client)
	if err != nil {
		if stats != nil {
			display.DisplayFilterStats(stats, s.Log)
		}
		return nil, nil, fmt.Errorf("gagal mendapatkan daftar database: %w", err)
	}

	// Simpan excluded databases untuk metadata (khusus untuk mode 'all')
	if s.BackupDBOptions.Mode == consts.ModeAll && stats != nil {
		s.excludedDatabases = stats.ExcludedDatabases
		s.Log.Infof("Menyimpan %d excluded databases untuk metadata", len(s.excludedDatabases))
		if len(s.excludedDatabases) > 0 {
			s.Log.Debugf("Excluded databases: %v", s.excludedDatabases)
		}
	}

	if len(dbFiltered) == 0 {
		display.DisplayFilterStats(stats, s.Log)
		ui.PrintError("Tidak ada database yang tersedia setelah filtering!")
		s.displayFilterWarnings(stats)
		return nil, nil, fmt.Errorf("tidak ada database tersedia untuk backup setelah filtering")
	}

	// Generate output directory dan filename
	// Untuk mode single: dbFiltered = [database_yang_dipilih]
	// Untuk mode primary/secondary: dbFiltered = [database_yang_dipilih, companion databases]
	dbFiltered, err = s.generateBackupPaths(ctx, client, dbFiltered)
	if err != nil {
		return nil, nil, err
	}

	if !showOptions {
		if proceed, askErr := display.NewOptionsDisplayer(s.BackupDBOptions).Display(); askErr != nil {
			return nil, nil, askErr
		} else if !proceed {
			return nil, nil, validation.ErrUserCancelled
		}
	}

	success = true
	return client, dbFiltered, nil
}
