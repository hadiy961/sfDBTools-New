package parsing

import (
	restoremodel "sfdbtools/internal/app/restore/model"
	resolver "sfdbtools/internal/cli/resolver"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"

	"github.com/spf13/cobra"
)

// PopulateProfileFlags membaca flag profile dan mengupdate struct.
func PopulateProfileFlags(cmd *cobra.Command, opts *domain.ProfileInfo) error {
	if v := resolver.GetStringFlagOrEnv(cmd, "profile", consts.ENV_SOURCE_PROFILE); v != "" {
		opts.Path = v
	}
	if v, err := resolver.GetSecretStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY); err != nil {
		return err
	} else if v != "" {
		opts.EncryptionKey = v
	}
	return nil
}

// PopulateEncryptionFlags membaca flag encryption dan mengupdate struct.
func PopulateEncryptionFlags(cmd *cobra.Command, opts *domain.EncryptionOptions) error {
	if v, err := resolver.GetSecretStringFlagOrEnv(cmd, "backup-key", consts.ENV_BACKUP_ENCRYPTION_KEY); err != nil {
		return err
	} else if v != "" {
		opts.Key = v
		opts.Enabled = true
	}
	return nil
}

// PopulateFilterFlags membaca flag filter dan mengupdate struct.
func PopulateFilterFlags(cmd *cobra.Command, opts *domain.FilterOptions) {
	if v := resolver.GetStringArrayFlagOrEnv(cmd, "db", ""); len(v) > 0 {
		opts.IncludeDatabases = v
	}
	if v := resolver.GetStringFlagOrEnv(cmd, "db-file", ""); v != "" {
		opts.IncludeFile = v
	}

	// Backup filters (default mengikuti config; override via flag jika ada)
	// Catatan: exclude-db/exclude-db-file tidak diekspos sebagai flag untuk backup.
	if cmd.Flags().Lookup("exclude-system") != nil {
		opts.ExcludeSystem = resolver.GetBoolFlagOrEnv(cmd, "exclude-system", "")
	}
	if cmd.Flags().Lookup("exclude-data") != nil {
		opts.ExcludeData = resolver.GetBoolFlagOrEnv(cmd, "exclude-data", "")
	}
	if cmd.Flags().Lookup("exclude-empty") != nil {
		opts.ExcludeEmpty = resolver.GetBoolFlagOrEnv(cmd, "exclude-empty", "")
	}
}

// -------------------- restore helpers --------------------

// PopulateTargetProfileFlags membaca flag profile target (restore) dan mengupdate struct.
func PopulateTargetProfileFlags(cmd *cobra.Command, opts *domain.ProfileInfo) error {
	if v := resolver.GetStringFlagOrEnv(cmd, "profile", consts.ENV_TARGET_PROFILE); v != "" {
		opts.Path = v
	}
	if v, err := resolver.GetSecretStringFlagOrEnv(cmd, "profile-key", consts.ENV_TARGET_PROFILE_KEY); err != nil {
		return err
	} else if v != "" {
		opts.EncryptionKey = v
	}
	return nil
}

// PopulateRestoreEncryptionKey membaca encryption key untuk decrypt backup file.
func PopulateRestoreEncryptionKey(cmd *cobra.Command, key *string) error {
	if v, err := resolver.GetSecretStringFlagOrEnv(cmd, "encryption-key", consts.ENV_BACKUP_ENCRYPTION_KEY); err != nil {
		return err
	} else if v != "" {
		*key = v
	}
	return nil
}

// PopulateRestoreSafetyFlags membaca flag safety umum untuk restore.
// Catatan: parameter 'force' adalah field model yang dipakai untuk menandai mode non-interaktif.
// Setelah refactor, mode non-interaktif dipicu oleh --skip-confirm (pengganti --force).
func PopulateRestoreSafetyFlags(cmd *cobra.Command, dropTarget, skipBackup, dryRun, force *bool) {
	if cmd.Flags().Changed("drop-target") {
		*dropTarget = resolver.GetBoolFlagOrEnv(cmd, "drop-target", "")
	}
	if cmd.Flags().Changed("skip-backup") {
		*skipBackup = resolver.GetBoolFlagOrEnv(cmd, "skip-backup", "")
	}
	if cmd.Flags().Changed("dry-run") {
		*dryRun = resolver.GetBoolFlagOrEnv(cmd, "dry-run", "")
	}
	if cmd.Flags().Changed("skip-confirm") {
		*force = resolver.GetBoolFlagOrEnv(cmd, "skip-confirm", "")
	}
}

// PopulateRestoreTicket membaca flag ticket.
func PopulateRestoreTicket(cmd *cobra.Command, ticket *string) {
	if v := resolver.GetStringFlagOrEnv(cmd, "ticket", ""); v != "" {
		*ticket = v
	}
}

// PopulateRestoreBackupDir membaca flag backup-dir ke RestoreBackupOptions.
func PopulateRestoreBackupDir(cmd *cobra.Command, opts *restoremodel.RestoreBackupOptions) {
	if v := resolver.GetStringFlagOrEnv(cmd, "backup-dir", ""); v != "" {
		opts.OutputDir = v
	}
}

// PopulateRestoreGrantsFlags membaca flag grants-file dan skip-grants.
func PopulateRestoreGrantsFlags(cmd *cobra.Command, grantsFile *string, skipGrants *bool) {
	if v := resolver.GetStringFlagOrEnv(cmd, "grants-file", ""); v != "" {
		*grantsFile = v
	}
	if cmd.Flags().Changed("skip-grants") {
		*skipGrants = resolver.GetBoolFlagOrEnv(cmd, "skip-grants", "")
	}
}

// PopulateStopOnErrorFromContinueFlag mengatur StopOnError berbasis flag continue-on-error.
func PopulateStopOnErrorFromContinueFlag(cmd *cobra.Command, stopOnError *bool) {
	if cmd.Flags().Changed("continue-on-error") {
		*stopOnError = !resolver.GetBoolFlagOrEnv(cmd, "continue-on-error", "")
	}
}
