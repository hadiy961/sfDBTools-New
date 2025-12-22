// File : pkg/parsing/parsing_restore.go
// Deskripsi : Parsing functions untuk restore options
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-16

package parsing

import (
	"sfDBTools/internal/types"
	"sfDBTools/pkg/helper"

	"github.com/spf13/cobra"
)

// ParsingRestoreSingleOptions melakukan parsing opsi untuk restore single
func ParsingRestoreSingleOptions(cmd *cobra.Command) (types.RestoreSingleOptions, error) {
	opts := types.RestoreSingleOptions{
		DropTarget: true,  // Default true
		SkipBackup: false, // Default false
	}

	// Profile & key (target)
	PopulateTargetProfileFlags(cmd, &opts.Profile)

	// Encryption key untuk decrypt backup file
	PopulateRestoreEncryptionKey(cmd, &opts.EncryptionKey)

	// Safety flags
	PopulateRestoreSafetyFlags(cmd, &opts.DropTarget, &opts.SkipBackup, &opts.DryRun, &opts.Force)

	// File backup
	if v := helper.GetStringFlagOrEnv(cmd, "file", ""); v != "" {
		opts.File = v
	}

	// Ticket number
	PopulateRestoreTicket(cmd, &opts.Ticket)

	// Target database
	if v := helper.GetStringFlagOrEnv(cmd, "target-db", ""); v != "" {
		opts.TargetDB = v
	}

	// Grants
	PopulateRestoreGrantsFlags(cmd, &opts.GrantsFile, &opts.SkipGrants)

	// Backup options untuk pre-restore backup
	opts.BackupOptions = &types.RestoreBackupOptions{}

	// Backup directory
	PopulateRestoreBackupDir(cmd, opts.BackupOptions)

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

	// Profile & key (target)
	PopulateTargetProfileFlags(cmd, &opts.Profile)

	// Encryption key untuk decrypt backup file
	PopulateRestoreEncryptionKey(cmd, &opts.EncryptionKey)

	// Safety flags (force handled khusus di bawah)
	PopulateRestoreSafetyFlags(cmd, &opts.DropTarget, &opts.SkipBackup, &opts.DryRun, &opts.Force)

	// File backup primary
	if v := helper.GetStringFlagOrEnv(cmd, "file", ""); v != "" {
		opts.File = v
	}

	// Companion file (dmart)
	if v := helper.GetStringFlagOrEnv(cmd, "companion-file", ""); v != "" {
		opts.CompanionFile = v
	}

	// Ticket number
	PopulateRestoreTicket(cmd, &opts.Ticket)

	// Target database
	if v := helper.GetStringFlagOrEnv(cmd, "target-db", ""); v != "" {
		opts.TargetDB = v
	}

	// Grants
	PopulateRestoreGrantsFlags(cmd, &opts.GrantsFile, &opts.SkipGrants)

	// Force mode special: jika force, maka skip konfirmasi.
	if opts.Force {
		opts.ConfirmIfNotExists = false
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
	PopulateRestoreBackupDir(cmd, opts.BackupOptions)

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
	PopulateTargetProfileFlags(cmd, &opts.Profile)
	if v := helper.GetStringFlagOrEnv(cmd, "file", ""); v != "" {
		opts.File = v
	}
	PopulateRestoreEncryptionKey(cmd, &opts.EncryptionKey)

	// 2. Safety Flags
	PopulateRestoreSafetyFlags(cmd, &opts.DropTarget, &opts.SkipBackup, &opts.DryRun, &opts.Force)
	PopulateStopOnErrorFromContinueFlag(cmd, &opts.StopOnError)

	// 4. Ticket & Backup Dir
	PopulateRestoreTicket(cmd, &opts.Ticket)
	opts.BackupOptions = &types.RestoreBackupOptions{}
	PopulateRestoreBackupDir(cmd, opts.BackupOptions)

	return opts, nil
}
