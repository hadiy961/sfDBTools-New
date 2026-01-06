package parsing

import (
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/domain"
	"sfdbtools/pkg/consts"
	"sfdbtools/pkg/helper"

	"github.com/spf13/cobra"
)

// PopulateProfileFlags membaca flag profile dan mengupdate struct.
func PopulateProfileFlags(cmd *cobra.Command, opts *domain.ProfileInfo) error {
	if v := helper.GetStringFlagOrEnv(cmd, "profile", consts.ENV_SOURCE_PROFILE); v != "" {
		opts.Path = v
	}
	if v, err := helper.GetSecretStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY); err != nil {
		return err
	} else if v != "" {
		opts.EncryptionKey = v
	}
	return nil
}

// PopulateEncryptionFlags membaca flag encryption dan mengupdate struct.
func PopulateEncryptionFlags(cmd *cobra.Command, opts *domain.EncryptionOptions) error {
	if v, err := helper.GetSecretStringFlagOrEnv(cmd, "backup-key", consts.ENV_BACKUP_ENCRYPTION_KEY); err != nil {
		return err
	} else if v != "" {
		opts.Key = v
		opts.Enabled = true
	}
	return nil
}

// PopulateFilterFlags membaca flag filter dan mengupdate struct.
func PopulateFilterFlags(cmd *cobra.Command, opts *domain.FilterOptions) {
	if v := helper.GetStringArrayFlagOrEnv(cmd, "db", ""); len(v) > 0 {
		opts.IncludeDatabases = v
	}
	if v := helper.GetStringFlagOrEnv(cmd, "db-file", ""); v != "" {
		opts.IncludeFile = v
	}

	// Backup filters (default mengikuti config; override via flag jika ada)
	// Catatan: exclude-db/exclude-db-file tidak diekspos sebagai flag untuk backup.
	if cmd.Flags().Lookup("exclude-system") != nil {
		opts.ExcludeSystem = helper.GetBoolFlagOrEnv(cmd, "exclude-system", "")
	}
	if cmd.Flags().Lookup("exclude-data") != nil {
		opts.ExcludeData = helper.GetBoolFlagOrEnv(cmd, "exclude-data", "")
	}
	if cmd.Flags().Lookup("exclude-empty") != nil {
		opts.ExcludeEmpty = helper.GetBoolFlagOrEnv(cmd, "exclude-empty", "")
	}
}

// -------------------- restore helpers --------------------

// PopulateTargetProfileFlags membaca flag profile target (restore) dan mengupdate struct.
func PopulateTargetProfileFlags(cmd *cobra.Command, opts *domain.ProfileInfo) error {
	if v := helper.GetStringFlagOrEnv(cmd, "profile", consts.ENV_TARGET_PROFILE); v != "" {
		opts.Path = v
	}
	if v, err := helper.GetSecretStringFlagOrEnv(cmd, "profile-key", consts.ENV_TARGET_PROFILE_KEY); err != nil {
		return err
	} else if v != "" {
		opts.EncryptionKey = v
	}
	return nil
}

// PopulateRestoreEncryptionKey membaca encryption key untuk decrypt backup file.
func PopulateRestoreEncryptionKey(cmd *cobra.Command, key *string) error {
	if v, err := helper.GetSecretStringFlagOrEnv(cmd, "encryption-key", consts.ENV_BACKUP_ENCRYPTION_KEY); err != nil {
		return err
	} else if v != "" {
		*key = v
	}
	return nil
}

// PopulateRestoreSafetyFlags membaca flag safety umum untuk restore.
func PopulateRestoreSafetyFlags(cmd *cobra.Command, dropTarget, skipBackup, dryRun, force *bool) {
	if cmd.Flags().Changed("drop-target") {
		*dropTarget = helper.GetBoolFlagOrEnv(cmd, "drop-target", "")
	}
	if cmd.Flags().Changed("skip-backup") {
		*skipBackup = helper.GetBoolFlagOrEnv(cmd, "skip-backup", "")
	}
	if cmd.Flags().Changed("dry-run") {
		*dryRun = helper.GetBoolFlagOrEnv(cmd, "dry-run", "")
	}
	if cmd.Flags().Changed("force") {
		*force = helper.GetBoolFlagOrEnv(cmd, "force", "")
	}
}

// PopulateRestoreTicket membaca flag ticket.
func PopulateRestoreTicket(cmd *cobra.Command, ticket *string) {
	if v := helper.GetStringFlagOrEnv(cmd, "ticket", ""); v != "" {
		*ticket = v
	}
}

// PopulateRestoreBackupDir membaca flag backup-dir ke RestoreBackupOptions.
func PopulateRestoreBackupDir(cmd *cobra.Command, opts *restoremodel.RestoreBackupOptions) {
	if v := helper.GetStringFlagOrEnv(cmd, "backup-dir", ""); v != "" {
		opts.OutputDir = v
	}
}

// PopulateRestoreGrantsFlags membaca flag grants-file dan skip-grants.
func PopulateRestoreGrantsFlags(cmd *cobra.Command, grantsFile *string, skipGrants *bool) {
	if v := helper.GetStringFlagOrEnv(cmd, "grants-file", ""); v != "" {
		*grantsFile = v
	}
	if cmd.Flags().Changed("skip-grants") {
		*skipGrants = helper.GetBoolFlagOrEnv(cmd, "skip-grants", "")
	}
}

// PopulateStopOnErrorFromContinueFlag mengatur StopOnError berbasis flag continue-on-error.
func PopulateStopOnErrorFromContinueFlag(cmd *cobra.Command, stopOnError *bool) {
	if cmd.Flags().Changed("continue-on-error") {
		*stopOnError = !helper.GetBoolFlagOrEnv(cmd, "continue-on-error", "")
	}
}
