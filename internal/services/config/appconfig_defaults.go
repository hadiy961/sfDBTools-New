// File : internal/services/config/appconfig_defaults.go
// Deskripsi : Default config + auto-init config file (zero-config first run)
// Author : Hadiyatna Muflihun
// Tanggal : 2 Januari 2026
// Last Modified : 20 Januari 2026
package appconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"sfdbtools/internal/app/version"
)

const (
	HardcodedAppName    = "sfdbtools"
	HardcodedAuthor     = "Hadiyatna Muflihun"
	HardcodedClientCode = "dataon"
)

func defaultConfigForPath(configPath string) *Config {
	baseDir := filepath.Dir(configPath)

	cfg := &Config{}
	cfg.General.AppName = HardcodedAppName
	cfg.General.Author = HardcodedAuthor
	cfg.General.ClientCode = HardcodedClientCode
	cfg.General.Locale.Timezone = "Asia/Jakarta"
	cfg.General.Locale.DateFormat = "2006-01-02"
	cfg.General.Locale.TimeFormat = "15:04:05"
	cfg.General.Version = version.Version

	cfg.Backup.Compression.Enabled = true
	cfg.Backup.Compression.Type = "zstd"
	cfg.Backup.Compression.Level = 5

	cfg.Backup.Exclude.SystemDatabases = true
	cfg.Backup.Exclude.User = false
	cfg.Backup.Exclude.Data = false
	cfg.Backup.Exclude.Empty = false

	cfg.Backup.Encryption.Enabled = true

	cfg.Backup.Output.BaseDirectory = "backup"
	cfg.Backup.Output.FilePermissions = "0600"
	cfg.Backup.Output.MetadataPermissions = "0600"
	cfg.Backup.Output.Structure.CreateSubdirs = true
	cfg.Backup.Output.Structure.Pattern = "{year}{month}{day}/"
	cfg.Backup.Output.SaveBackupInfo = true

	// Scheduler default: tidak ada job.
	// Job scheduler hanya aktif jika user mendefinisikan backup.scheduler.jobs di config.
	cfg.Backup.Scheduler.Jobs = nil

	cfg.Log.Level = "info"
	cfg.Log.Format = "text"
	cfg.Log.Output.Console.Enabled = true
	cfg.Log.Output.File.Enabled = false
	cfg.Log.Output.File.Dir = "logs"
	cfg.Log.Timezone = "Asia/Jakarta"

	cfg.Profile.Connection.Timeout = "15s"

	cfg.ConfigDir.DatabaseProfile = filepath.Join(baseDir, "config", "db_profile")
	cfg.Script.BundleOutputDir = filepath.Join(baseDir, "scripts")

	return cfg
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func writeDefaultConfig(configPath string) error {
	cfg := defaultConfigForPath(configPath)
	if err := ensureDir(filepath.Dir(configPath)); err != nil {
		return err
	}
	if err := ensureDir(filepath.Dir(cfg.ConfigDir.DatabaseProfile)); err != nil {
		return err
	}
	if cfg.Script.BundleOutputDir != "" {
		if err := ensureDir(cfg.Script.BundleOutputDir); err != nil {
			return err
		}
	}

	data, err := MarshalYAML(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return err
	}
	return nil
}

func userDefaultConfigPath() (string, error) {
	// Prefer XDG_CONFIG_HOME if set.
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "sfdbtools", "config.yaml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("gagal membaca home dir: %w", err)
	}
	return filepath.Join(home, ".config", "sfdbtools", "config.yaml"), nil
}
