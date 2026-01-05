package setup

import (
	"fmt"
	"strings"

	"sfDBTools/internal/services/config"
	"sfDBTools/internal/services/log"
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

// isInteractiveMode adalah source-of-truth tunggal untuk mode-mode backup yang
// memperbolehkan flow interaktif (pemilihan profile dan edit opsi).
func isInteractiveMode(mode string) bool {
	switch mode {
	case consts.ModeSingle, consts.ModePrimary, consts.ModeSecondary, consts.ModeCombined, consts.ModeAll, consts.ModeSeparated:
		return true
	default:
		return false
	}
}

// CheckAndSelectConfigFile loads a source profile (interactive if allowed) and stores it into options.
func (s *Setup) CheckAndSelectConfigFile() error {
	if s.Options.NonInteractive {
		if strings.TrimSpace(s.Options.Profile.Path) == "" {
			return fmt.Errorf("profile wajib diisi pada mode non-interaktif (--quiet): gunakan --profile")
		}
		if strings.TrimSpace(s.Options.Profile.EncryptionKey) == "" {
			return fmt.Errorf("profile-key wajib diisi pada mode non-interaktif (--quiet): gunakan --profile-key atau env %s", consts.ENV_SOURCE_PROFILE_KEY)
		}
	}

	allowInteractive := isInteractiveMode(s.Options.Mode) && !s.Options.NonInteractive && s.Options.Profile.Path == ""
	profile, err := profilehelper.LoadSourceProfile(
		s.Config.ConfigDir.DatabaseProfile,
		s.Options.Profile.Path,
		s.Options.Profile.EncryptionKey,
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

	if strings.TrimSpace(s.Options.Ticket) == "" {
		if s.Options.NonInteractive {
			return fmt.Errorf("ticket wajib diisi pada mode non-interaktif (--quiet): gunakan --ticket")
		}
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
		if strings.TrimSpace(s.Options.Encryption.Key) == "" {
			return fmt.Errorf("encryption diaktifkan tapi backup key kosong: gunakan --backup-key atau ENV %s (atau nonaktifkan encryption)", consts.ENV_BACKUP_ENCRYPTION_KEY)
		}
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
