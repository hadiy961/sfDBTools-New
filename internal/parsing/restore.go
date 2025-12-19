// File : pkg/parsing/parsing_restore.go
// Deskripsi : Parsing functions untuk restore options
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-16

package parsing

import (
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"

	"github.com/spf13/cobra"
)

// ParsingRestoreSingleOptions melakukan parsing opsi untuk restore single
func ParsingRestoreSingleOptions(cmd *cobra.Command) (types.RestoreSingleOptions, error) {
	opts := types.RestoreSingleOptions{
		DropTarget: true,  // Default true
		SkipBackup: false, // Default false
	}

	// Profile & key
	if v := helper.GetStringFlagOrEnv(cmd, "profile", consts.ENV_TARGET_PROFILE); v != "" {
		opts.Profile.Path = v
	}
	if v := helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_TARGET_PROFILE_KEY); v != "" {
		opts.Profile.EncryptionKey = v
	}

	// Encryption key untuk decrypt backup file
	if v := helper.GetStringFlagOrEnv(cmd, "encryption-key", consts.ENV_BACKUP_ENCRYPTION_KEY); v != "" {
		opts.EncryptionKey = v
	}

	// Drop target database
	if cmd.Flags().Changed("drop-target") {
		opts.DropTarget = helper.GetBoolFlagOrEnv(cmd, "drop-target", "")
	}

	// Skip backup
	if cmd.Flags().Changed("skip-backup") {
		opts.SkipBackup = helper.GetBoolFlagOrEnv(cmd, "skip-backup", "")
	}

	// File backup
	if v := helper.GetStringFlagOrEnv(cmd, "file", ""); v != "" {
		opts.File = v
	}

	// Ticket number
	if v := helper.GetStringFlagOrEnv(cmd, "ticket", ""); v != "" {
		opts.Ticket = v
	}

	// Target database
	if v := helper.GetStringFlagOrEnv(cmd, "target-db", ""); v != "" {
		opts.TargetDB = v
	}

	// User grants file
	if v := helper.GetStringFlagOrEnv(cmd, "grants-file", ""); v != "" {
		opts.GrantsFile = v
	}

	// Skip grants
	if cmd.Flags().Changed("skip-grants") {
		opts.SkipGrants = helper.GetBoolFlagOrEnv(cmd, "skip-grants", "")
	}

	// Dry-run mode
	if cmd.Flags().Changed("dry-run") {
		opts.DryRun = helper.GetBoolFlagOrEnv(cmd, "dry-run", "")
	}

	// Force mode
	if cmd.Flags().Changed("force") {
		opts.Force = helper.GetBoolFlagOrEnv(cmd, "force", "")
	}

	// Backup options untuk pre-restore backup
	opts.BackupOptions = &types.RestoreBackupOptions{}

	// Backup directory
	if v := helper.GetStringFlagOrEnv(cmd, "backup-dir", ""); v != "" {
		opts.BackupOptions.OutputDir = v
	}

	return opts, nil
}

// ParsingRestorePrimaryOptions melakukan parsing opsi untuk restore primary
func ParsingRestorePrimaryOptions(cmd *cobra.Command) (types.RestorePrimaryOptions, error) {
	opts := types.RestorePrimaryOptions{
		DropTarget:         true,  // Default true
		SkipBackup:         false, // Default false
		IncludeDmart:       true,  // Default true
		AutoDetectDmart:    true,  // Default true
		ConfirmIfNotExists: true,  // Default true
	}

	// Profile & key
	if v := helper.GetStringFlagOrEnv(cmd, "profile", consts.ENV_TARGET_PROFILE); v != "" {
		opts.Profile.Path = v
	}
	if v := helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_TARGET_PROFILE_KEY); v != "" {
		opts.Profile.EncryptionKey = v
	}

	// Encryption key untuk decrypt backup file
	if v := helper.GetStringFlagOrEnv(cmd, "encryption-key", consts.ENV_BACKUP_ENCRYPTION_KEY); v != "" {
		opts.EncryptionKey = v
	}

	// Drop target database
	if cmd.Flags().Changed("drop-target") {
		opts.DropTarget = helper.GetBoolFlagOrEnv(cmd, "drop-target", "")
	}

	// Skip backup
	if cmd.Flags().Changed("skip-backup") {
		opts.SkipBackup = helper.GetBoolFlagOrEnv(cmd, "skip-backup", "")
	}

	// File backup primary
	if v := helper.GetStringFlagOrEnv(cmd, "file", ""); v != "" {
		opts.File = v
	}

	// Companion file (dmart)
	if v := helper.GetStringFlagOrEnv(cmd, "companion-file", ""); v != "" {
		opts.CompanionFile = v
	}

	// Ticket number
	if v := helper.GetStringFlagOrEnv(cmd, "ticket", ""); v != "" {
		opts.Ticket = v
	}

	// Target database
	if v := helper.GetStringFlagOrEnv(cmd, "target-db", ""); v != "" {
		opts.TargetDB = v
	}

	// User grants file
	if v := helper.GetStringFlagOrEnv(cmd, "grants-file", ""); v != "" {
		opts.GrantsFile = v
	} // Skip grants
	if cmd.Flags().Changed("skip-grants") {
		opts.SkipGrants = helper.GetBoolFlagOrEnv(cmd, "skip-grants", "")
	}
	// Dry-run mode
	if cmd.Flags().Changed("dry-run") {
		opts.DryRun = helper.GetBoolFlagOrEnv(cmd, "dry-run", "")
	}
	// Force mode
	if cmd.Flags().Changed("force") {
		opts.Force = helper.GetBoolFlagOrEnv(cmd, "force", "")
		if opts.Force {
			opts.ConfirmIfNotExists = false
		}
	}

	// Include dmart
	if cmd.Flags().Changed("include-dmart") {
		opts.IncludeDmart = helper.GetBoolFlagOrEnv(cmd, "include-dmart", "")
	}

	// Auto-detect dmart
	if cmd.Flags().Changed("auto-detect-dmart") {
		opts.AutoDetectDmart = helper.GetBoolFlagOrEnv(cmd, "auto-detect-dmart", "")
	}

	// Confirm if not exists
	if cmd.Flags().Changed("skip-confirm") {
		opts.ConfirmIfNotExists = !helper.GetBoolFlagOrEnv(cmd, "skip-confirm", "")
	}

	// Backup options untuk pre-restore backup
	opts.BackupOptions = &types.RestoreBackupOptions{}

	// Backup directory
	if v := helper.GetStringFlagOrEnv(cmd, "backup-dir", ""); v != "" {
		opts.BackupOptions.OutputDir = v
	}

	return opts, nil
}

// ParsingRestoreAllOptions melakukan parsing opsi untuk restore all
func ParsingRestoreAllOptions(cmd *cobra.Command) (types.RestoreAllOptions, error) {
	opts := types.RestoreAllOptions{
		SkipBackup:    false,
		SkipSystemDBs: true, // Default true demi keamanan
		StopOnError:   true,
	}

	// 1. Basic configs (Profile, Keys, File)
	if v := helper.GetStringFlagOrEnv(cmd, "profile", consts.ENV_TARGET_PROFILE); v != "" {
		opts.Profile.Path = v
	}
	if v := helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_TARGET_PROFILE_KEY); v != "" {
		opts.Profile.EncryptionKey = v
	}
	if v := helper.GetStringFlagOrEnv(cmd, "file", ""); v != "" {
		opts.File = v
	}
	if v := helper.GetStringFlagOrEnv(cmd, "encryption-key", consts.ENV_BACKUP_ENCRYPTION_KEY); v != "" {
		opts.EncryptionKey = v
	}

	// 2. Safety Flags
	if cmd.Flags().Changed("skip-backup") {
		opts.SkipBackup = helper.GetBoolFlagOrEnv(cmd, "skip-backup", "")
	}
	if cmd.Flags().Changed("dry-run") {
		opts.DryRun = helper.GetBoolFlagOrEnv(cmd, "dry-run", "")
	}
	if cmd.Flags().Changed("force") {
		opts.Force = helper.GetBoolFlagOrEnv(cmd, "force", "")
	}
	if cmd.Flags().Changed("continue-on-error") {
		opts.StopOnError = !helper.GetBoolFlagOrEnv(cmd, "continue-on-error", "")
	}
	if cmd.Flags().Changed("drop-target") {
		opts.DropTarget = helper.GetBoolFlagOrEnv(cmd, "drop-target", "")
	}

	// 4. Ticket & Backup Dir
	if v := helper.GetStringFlagOrEnv(cmd, "ticket", ""); v != "" {
		opts.Ticket = v
	}
	opts.BackupOptions = &types.RestoreBackupOptions{}
	if v := helper.GetStringFlagOrEnv(cmd, "backup-dir", ""); v != "" {
		opts.BackupOptions.OutputDir = v
	}

	return opts, nil
}
