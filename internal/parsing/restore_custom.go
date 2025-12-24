// File : internal/parsing/restore_custom.go
// Deskripsi : Parsing flags untuk restore custom (SFCola account detail)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-24

package parsing

import (
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
)

// ParsingRestoreCustomOptions melakukan parsing opsi untuk restore custom.
// Catatan: detail account (database/user/password) diprompt secara interaktif saat setup session.
func ParsingRestoreCustomOptions(cmd *cobra.Command) (types.RestoreCustomOptions, error) {
	opts := types.RestoreCustomOptions{
		DropTarget:  true,
		SkipBackup:  false,
		StopOnError: true,
	}

	// Profile & key (target)
	PopulateTargetProfileFlags(cmd, &opts.Profile)

	// Encryption key untuk decrypt backup file
	PopulateRestoreEncryptionKey(cmd, &opts.EncryptionKey)

	// Safety flags
	PopulateRestoreSafetyFlags(cmd, &opts.DropTarget, &opts.SkipBackup, &opts.DryRun, &opts.Force)
	PopulateStopOnErrorFromContinueFlag(cmd, &opts.StopOnError)

	// Ticket
	PopulateRestoreTicket(cmd, &opts.Ticket)

	// Backup options untuk pre-restore backup
	opts.BackupOptions = &types.RestoreBackupOptions{}
	PopulateRestoreBackupDir(cmd, opts.BackupOptions)

	return opts, nil
}
