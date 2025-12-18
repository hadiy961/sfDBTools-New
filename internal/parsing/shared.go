package parsing

import (
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"

	"github.com/spf13/cobra"
)

// PopulateProfileFlags membaca flag profile dan mengupdate struct.
func PopulateProfileFlags(cmd *cobra.Command, opts *types.ProfileInfo) {
	if v := helper.GetStringFlagOrEnv(cmd, "profile", consts.ENV_SOURCE_PROFILE); v != "" {
		opts.Path = v
	}
	if v := helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY); v != "" {
		opts.EncryptionKey = v
	}
}

// PopulateEncryptionFlags membaca flag encryption dan mengupdate struct.
func PopulateEncryptionFlags(cmd *cobra.Command, opts *types_backup.EncryptionOptions) {
	if v := helper.GetStringFlagOrEnv(cmd, "encryption-key", consts.ENV_BACKUP_ENCRYPTION_KEY); v != "" {
		opts.Key = v
		opts.Enabled = true
	}
}

// PopulateCompressionFlags membaca flag compression dan mengupdate struct.
func PopulateCompressionFlags(cmd *cobra.Command, opts *types_backup.CompressionOptions) {
	if v := helper.GetStringFlagOrEnv(cmd, "compress-type", ""); v != "" {
		if v == "none" {
			opts.Type = v
			opts.Enabled = false
		} else {
			opts.Type = v
			opts.Enabled = true
		}
	}
	if v := helper.GetIntFlagOrEnv(cmd, "compress-level", ""); v != 0 {
		opts.Level = v
	}
}

// PopulateFilterFlags membaca flag filter dan mengupdate struct.
func PopulateFilterFlags(cmd *cobra.Command, opts *types.FilterOptions) {
	if cmd.Flags().Changed("exclude-system") {
		opts.ExcludeSystem = helper.GetBoolFlagOrEnv(cmd, "exclude-system", "")
	}

	if v := helper.GetStringArrayFlagOrEnv(cmd, "exclude-db", ""); len(v) > 0 {
		opts.ExcludeDatabases = v
	}
	if v := helper.GetStringFlagOrEnv(cmd, "exclude-db-file", ""); v != "" {
		opts.ExcludeDBFile = v
	}
	// Cek alias
	if opts.ExcludeDBFile == "" {
		if v := helper.GetStringFlagOrEnv(cmd, "exclude-file", ""); v != "" {
			opts.ExcludeDBFile = v
		}
	}

	if v := helper.GetStringArrayFlagOrEnv(cmd, "db", ""); len(v) > 0 {
		opts.IncludeDatabases = v
	}
	if v := helper.GetStringFlagOrEnv(cmd, "db-file", ""); v != "" {
		opts.IncludeFile = v
	}

	// Exclude options (untuk backup all)
	if cmd.Flags().Changed("exclude-data") {
		opts.ExcludeData = helper.GetBoolFlagOrEnv(cmd, "exclude-data", "")
	}
	if cmd.Flags().Changed("exclude-empty") {
		opts.ExcludeEmpty = helper.GetBoolFlagOrEnv(cmd, "exclude-empty", "")
	}
}
