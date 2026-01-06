// File : internal/cli/parsing/restore_custom.go
// Deskripsi : Parsing flags untuk restore custom (SFCola account detail)
// Author : Hadiyatna Muflihun
// Tanggal : 24 Desember 2025
// Last Modified : 6 Januari 2026

package parsing

import (
	restoremodel "sfdbtools/internal/app/restore/model"

	"github.com/spf13/cobra"
)

// ParsingRestoreCustomOptions melakukan parsing opsi untuk restore custom.
// Catatan: detail account (database/user/password) diprompt secara interaktif saat setup session.
func ParsingRestoreCustomOptions(cmd *cobra.Command) (restoremodel.RestoreCustomOptions, error) {
	opts := restoremodel.RestoreCustomOptions{
		DropTarget:  true,
		SkipBackup:  false,
		StopOnError: true,
	}

	// Profile & key (target)
	if err := PopulateTargetProfileFlags(cmd, &opts.Profile); err != nil {
		return restoremodel.RestoreCustomOptions{}, err
	}

	// Encryption key untuk decrypt backup file
	if err := PopulateRestoreEncryptionKey(cmd, &opts.EncryptionKey); err != nil {
		return restoremodel.RestoreCustomOptions{}, err
	}

	// Safety flags
	PopulateRestoreSafetyFlags(cmd, &opts.DropTarget, &opts.SkipBackup, &opts.DryRun, &opts.Force)
	PopulateStopOnErrorFromContinueFlag(cmd, &opts.StopOnError)

	// Ticket
	PopulateRestoreTicket(cmd, &opts.Ticket)

	// Backup options untuk pre-restore backup
	opts.BackupOptions = &restoremodel.RestoreBackupOptions{}
	PopulateRestoreBackupDir(cmd, opts.BackupOptions)

	return opts, nil
}
