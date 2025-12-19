// File : internal/parsing/restore_selection.go
// Deskripsi : Parsing flags untuk restore selection (CSV)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-19

package parsing

import (
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"

	"github.com/spf13/cobra"
)

// ParsingRestoreSelectionOptions melakukan parsing opsi untuk restore selection
func ParsingRestoreSelectionOptions(cmd *cobra.Command) (types.RestoreSelectionOptions, error) {
	opts := types.RestoreSelectionOptions{
		DropTarget:  true,
		SkipBackup:  false,
		StopOnError: true, // default stop pada error pertama
	}

	// Profile & key
	if v := helper.GetStringFlagOrEnv(cmd, "profile", consts.ENV_TARGET_PROFILE); v != "" {
		opts.Profile.Path = v
	}
	if v := helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_TARGET_PROFILE_KEY); v != "" {
		opts.Profile.EncryptionKey = v
	}

	// CSV source
	if v := helper.GetStringFlagOrEnv(cmd, "csv", ""); v != "" {
		opts.CSV = v
	}

	// Safety flags
	if cmd.Flags().Changed("drop-target") {
		opts.DropTarget = helper.GetBoolFlagOrEnv(cmd, "drop-target", "")
	}
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

	// Ticket
	if v := helper.GetStringFlagOrEnv(cmd, "ticket", ""); v != "" {
		opts.Ticket = v
	}

	// Backup options
	opts.BackupOptions = &types.RestoreBackupOptions{}
	if v := helper.GetStringFlagOrEnv(cmd, "backup-dir", ""); v != "" {
		opts.BackupOptions.OutputDir = v
	}

	return opts, nil
}
