package setup

import (
	"context"
	"fmt"

	"sfDBTools/internal/backup/display"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	profilehelper "sfDBTools/pkg/helper/profile"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

type PathGenerator func(ctx context.Context, client *database.Client, dbFiltered []string) ([]string, error)

// PrepareBackupSession runs the whole pre-backup preparation flow.
func (s *Setup) PrepareBackupSession(ctx context.Context, headerTitle string, showOptions bool, genPaths PathGenerator) (client *database.Client, dbFiltered []string, err error) {
	if headerTitle != "" {
		ui.Headers(headerTitle)
	}

	if err = s.CheckAndSelectConfigFile(); err != nil {
		return nil, nil, err
	}

	client, err = profilehelper.ConnectWithProfile(&s.Options.Profile, consts.DefaultInitialDatabase)
	if err != nil {
		return nil, nil, err
	}

	var success bool
	defer func() {
		if !success && client != nil {
			client.Close()
		}
	}()

	serverHostname, err := client.GetServerHostname(ctx)
	if err != nil {
		s.Log.Warnf("gagal mendapatkan hostname dari server: %v, menggunakan dari config", err)
		serverHostname = s.Options.Profile.DBInfo.Host
	} else {
		s.Options.Profile.DBInfo.HostName = serverHostname
		s.Log.Infof("menggunakan hostname dari server: %s", serverHostname)
	}

	var stats *types.FilterStats
	dbFiltered, stats, err = s.GetFilteredDatabases(ctx, client)
	if err != nil {
		if stats != nil {
			ui.DisplayFilterStats(stats, consts.FeatureBackup, s.Log)
		}
		return nil, nil, fmt.Errorf("gagal mendapatkan daftar database: %w", err)
	}

	if s.Options.Mode == consts.ModeAll && stats != nil && s.ExcludedDatabases != nil {
		*s.ExcludedDatabases = stats.ExcludedDatabases
		s.Log.Infof("Menyimpan %d excluded databases untuk metadata", len(*s.ExcludedDatabases))
		if len(*s.ExcludedDatabases) > 0 {
			s.Log.Debugf("Excluded databases: %v", *s.ExcludedDatabases)
		}
	}

	if len(dbFiltered) == 0 {
		ui.DisplayFilterStats(stats, consts.FeatureBackup, s.Log)
		ui.PrintError("Tidak ada database yang tersedia setelah filtering!")
		if stats != nil {
			s.DisplayFilterWarnings(stats)
		}
		return nil, nil, fmt.Errorf("tidak ada database tersedia untuk backup setelah filtering")
	}

	if genPaths == nil {
		return nil, nil, fmt.Errorf("path generator tidak tersedia")
	}
	// Generate output directory and filename (and expand dbFiltered for single/primary/secondary).
	dbFiltered, err = genPaths(ctx, client, dbFiltered)
	if err != nil {
		return nil, nil, err
	}

	// Validate hasil path generation - dbFiltered tidak boleh empty
	if len(dbFiltered) == 0 {
		return nil, nil, fmt.Errorf("path generation menghasilkan daftar database kosong")
	}

	if !showOptions {
		if proceed, askErr := display.NewOptionsDisplayer(s.Options).Display(); askErr != nil {
			return nil, nil, askErr
		} else if !proceed {
			return nil, nil, validation.ErrUserCancelled
		}
	}

	success = true
	return client, dbFiltered, nil
}
