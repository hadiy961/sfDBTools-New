package parsing

import (
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	defaultVal "sfDBTools/pkg/defaultval"
	"sfDBTools/pkg/helper"

	"github.com/spf13/cobra"
)

// ParsingBackupOptions melakukan parsing opsi untuk backup combined
func ParsingBackupOptions(cmd *cobra.Command, mode string) (types.BackupDBOptions, error) {
	// Mulai dari default untuk mode combined
	opts := defaultVal.DefaultBackupOptions(mode)
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

	// Compression
	opts.Compression.Enabled = helper.GetBoolFlagOrEnv(cmd, "compress", "")
	if v := helper.GetStringFlagOrEnv(cmd, "compression-type", ""); v != "" {
		opts.Compression.Type = v
	}
	if v := helper.GetIntFlagOrEnv(cmd, "compression-level", ""); v != 0 {
		opts.Compression.Level = v
	}

	// Encryption
	opts.Encryption.Enabled = helper.GetBoolFlagOrEnv(cmd, "encrypt", "")
	if v := helper.GetStringFlagOrEnv(cmd, "encryption-key", consts.ENV_BACKUP_ENCRYPTION_KEY); v != "" {
		opts.Encryption.Key = v
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

	// Mode
	opts.Mode = mode

	return opts, nil
}
