package parsing

import (
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/consts"
	defaultVal "sfDBTools/pkg/defaultval"
	"sfDBTools/pkg/helper"

	"github.com/spf13/cobra"
)

// ParsingBackupOptions melakukan parsing opsi untuk backup combined
func ParsingBackupOptions(cmd *cobra.Command, mode string) (types_backup.BackupDBOptions, error) {
	// Mulai dari default untuk mode combined
	opts := defaultVal.DefaultBackupOptions(mode)

	// Deteksi apakah ini command filter (untuk multi-select logic)
	// Command filter memiliki Use="filter", sedangkan all memiliki Use="all"
	isFilterCommand := cmd.Use == "filter"

	// Profile & key
	if v := helper.GetStringFlagOrEnv(cmd, "profile", consts.ENV_SOURCE_PROFILE); v != "" {
		opts.Profile.Path = v
	}
	if v := helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY); v != "" {
		opts.Encryption.Key = v
	}

	// Filters
	opts.Filter.ExcludeSystem = helper.GetBoolFlagOrEnv(cmd, "exclude-system", "")
	opts.Filter.ExcludeDatabases = helper.GetStringArrayFlagOrEnv(cmd, "exclude-db", "")
	opts.Filter.ExcludeDBFile = helper.GetStringFlagOrEnv(cmd, "exclude-db-file", "")
	opts.Filter.IncludeDatabases = helper.GetStringArrayFlagOrEnv(cmd, "db", "")
	opts.Filter.IncludeFile = helper.GetStringFlagOrEnv(cmd, "db-file", "")

	// Set flag untuk command filter agar bisa tampilkan multi-select jika tidak ada include/exclude
	// Ini digunakan di setup.go untuk menentukan apakah perlu multi-select atau tidak
	if isFilterCommand {
		opts.Filter.IsFilterCommand = true
	}

	// Compression - derive from compress-type (enabled unless type is "none" or empty)
	if v := helper.GetStringFlagOrEnv(cmd, "compress-type", ""); v != "" && v != "none" {
		opts.Compression.Type = v
		opts.Compression.Enabled = true
	} else if v == "none" {
		opts.Compression.Type = v
		opts.Compression.Enabled = false
	}
	if v := helper.GetIntFlagOrEnv(cmd, "compress-level", ""); v != 0 {
		opts.Compression.Level = v
	}

	// Encryption - derive from encryption-key (enabled if key is not empty)
	if v := helper.GetStringFlagOrEnv(cmd, "encryption-key", consts.ENV_BACKUP_ENCRYPTION_KEY); v != "" {
		opts.Encryption.Key = v
		opts.Encryption.Enabled = true
	}

	// Capture GTID (hanya untuk combined)
	if mode == "combined" {
		opts.CaptureGTID = helper.GetBoolFlagOrEnv(cmd, "capture-gtid", "")
	} else {
		opts.CaptureGTID = false
	}

	// Exclude User
	opts.ExcludeUser = helper.GetBoolFlagOrEnv(cmd, "exclude-user", "")

	// Dry Run
	opts.DryRun = helper.GetBoolFlagOrEnv(cmd, "dry-run", "")

	// Output Directory
	if v := helper.GetStringFlagOrEnv(cmd, "output-dir", ""); v != "" {
		opts.OutputDir = v
	}
	opts.Force = helper.GetBoolFlagOrEnv(cmd, "force", "")

	// Mode-specific options
	if mode == "single" {
		if v := helper.GetStringFlagOrEnv(cmd, "database", ""); v != "" {
			opts.DBName = v
		}
		if v := helper.GetStringFlagOrEnv(cmd, "filename", ""); v != "" {
			opts.File.Filename = v
		}
		opts.Filter.ExcludeData = helper.GetBoolFlagOrEnv(cmd, "exclude-data", "")
		opts.IncludeDmart = helper.GetBoolFlagOrEnv(cmd, "include-dmart", "")
		opts.IncludeTemp = helper.GetBoolFlagOrEnv(cmd, "include-temp", "")
		opts.IncludeArchive = helper.GetBoolFlagOrEnv(cmd, "include-archive", "")
	} else if mode == "primary" {
		// Mode primary sama seperti single, hanya tanpa --database flag
		if v := helper.GetStringFlagOrEnv(cmd, "filename", ""); v != "" {
			opts.File.Filename = v
		}
		opts.Filter.ExcludeData = helper.GetBoolFlagOrEnv(cmd, "exclude-data", "")
		opts.IncludeDmart = helper.GetBoolFlagOrEnv(cmd, "include-dmart", "")
		opts.IncludeTemp = helper.GetBoolFlagOrEnv(cmd, "include-temp", "")
		opts.IncludeArchive = helper.GetBoolFlagOrEnv(cmd, "include-archive", "")
	} else if mode == "secondary" {
		// Mode secondary sama seperti primary, hanya untuk database dengan suffix _secondary
		if v := helper.GetStringFlagOrEnv(cmd, "filename", ""); v != "" {
			opts.File.Filename = v
		}
		opts.Filter.ExcludeData = helper.GetBoolFlagOrEnv(cmd, "exclude-data", "")
	}

	// Mode
	opts.Mode = mode

	return opts, nil
}
