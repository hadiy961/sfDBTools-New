package setup

import (
	"fmt"

	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/fsops"
	profilehelper "sfDBTools/pkg/helper/profile"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
)

type Setup struct {
	Log               applog.Logger
	Config            *appconfig.Config
	Options           *types_backup.BackupDBOptions
	ExcludedDatabases *[]string
}

func New(log applog.Logger, cfg *appconfig.Config, opts *types_backup.BackupDBOptions, excluded *[]string) *Setup {
	return &Setup{Log: log, Config: cfg, Options: opts, ExcludedDatabases: excluded}
}

func allowInteractiveProfileSelection(mode string) bool {
	switch mode {
	case consts.ModeSingle, consts.ModePrimary, consts.ModeSecondary, consts.ModeCombined, consts.ModeAll, consts.ModeSeparated:
		return true
	default:
		return false
	}
}

// CheckAndSelectConfigFile loads a source profile (interactive if allowed) and stores it into options.
func (s *Setup) CheckAndSelectConfigFile() error {
	allowInteractive := allowInteractiveProfileSelection(s.Options.Mode) && s.Options.Profile.Path == ""
	profile, err := profilehelper.LoadSourceProfile(
		s.Config.ConfigDir.DatabaseProfile,
		s.Options.Profile.Path,
		s.Options.Encryption.Key,
		allowInteractive,
	)
	if err != nil {
		return fmt.Errorf("gagal load source profile: %w", err)
	}

	s.Options.Profile = *profile
	return nil
}

// SetupBackupExecution prepares common backup execution settings (ticket, output dir, logging).
func (s *Setup) SetupBackupExecution() error {
	ui.PrintSubHeader("Persiapan Eksekusi Backup")

	if s.Options.Ticket == "" {
		s.Log.Info("Ticket number tidak ditemukan, meminta input...")
		ticket, err := input.AskTicket(consts.FeatureBackup)
		if err != nil {
			return fmt.Errorf("gagal mendapatkan ticket number: %w", err)
		}
		s.Options.Ticket = ticket
		s.Log.Infof("Ticket number: %s", s.Options.Ticket)
	} else {
		s.Log.Infof("Ticket number: %s", s.Options.Ticket)
	}

	s.Log.Info("Membuat direktori output jika belum ada : " + s.Options.OutputDir)
	if err := fsops.CreateDirIfNotExist(s.Options.OutputDir); err != nil {
		return fmt.Errorf("gagal membuat direktori output: %w", err)
	}
	s.Log.Info("Direktori output siap: " + s.Options.OutputDir)

	if s.Options.Encryption.Enabled {
		s.Log.Info("Enkripsi AES-256-GCM diaktifkan untuk backup (kompatibel dengan OpenSSL)")
	} else {
		s.Log.Info("Enkripsi tidak diaktifkan, melewati langkah kunci enkripsi...")
	}

	if s.Options.Compression.Enabled {
		s.Log.Infof("Kompresi %s diaktifkan (level: %d)", s.Options.Compression.Type, s.Options.Compression.Level)
	} else {
		s.Log.Info("Kompresi tidak diaktifkan, melewati langkah kompresi...")
	}

	if s.Options.Filter.ExcludeData {
		s.Log.Info("Opsi exclude-data diaktifkan: hanya struktur database yang akan di-backup.")
	} else {
		s.Log.Info("Data database akan disertakan dalam backup.")
	}

	return nil
}
