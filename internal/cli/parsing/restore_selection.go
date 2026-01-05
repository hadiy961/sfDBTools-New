// File : internal/cli/parsing/restore_selection.go
// Deskripsi : Parsing flags untuk restore selection (CSV)
// Author : Hadiyatna Muflihun
// Tanggal : 19 Desember 2025

package parsing

import (
	restoremodel "sfDBTools/internal/app/restore/model"
	"sfDBTools/pkg/helper"

	"github.com/spf13/cobra"
)

// ParsingRestoreSelectionOptions melakukan parsing opsi untuk restore selection
func ParsingRestoreSelectionOptions(cmd *cobra.Command) (restoremodel.RestoreSelectionOptions, error) {
	opts := restoremodel.RestoreSelectionOptions{
		DropTarget:  true,
		SkipBackup:  false,
		StopOnError: true, // default stop pada error pertama
	}

	// Profile & key (target)
	PopulateTargetProfileFlags(cmd, &opts.Profile)

	// CSV source
	if v := helper.GetStringFlagOrEnv(cmd, "csv", ""); v != "" {
		opts.CSV = v
	}

	// Safety flags
	PopulateRestoreSafetyFlags(cmd, &opts.DropTarget, &opts.SkipBackup, &opts.DryRun, &opts.Force)
	PopulateStopOnErrorFromContinueFlag(cmd, &opts.StopOnError)

	// Ticket
	PopulateRestoreTicket(cmd, &opts.Ticket)

	// Backup options
	opts.BackupOptions = &restoremodel.RestoreBackupOptions{}
	PopulateRestoreBackupDir(cmd, opts.BackupOptions)

	return opts, nil
}
