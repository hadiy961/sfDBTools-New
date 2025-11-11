// File : cmd/cmd_restore/cmd_restore_multi.go
// Deskripsi : Command untuk restore multiple databases dari direktori backup files
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-10

package cmdrestore

import (
	"context"
	"fmt"
	"os"
	"sfDBTools/internal/restore"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/flags"

	"github.com/spf13/cobra"
)

// CmdRestoreMulti adalah command untuk restore multiple databases dari direktori backup files
var CmdRestoreMulti = &cobra.Command{
	Use:   "multi",
	Short: "Restore multiple databases dari direktori backup files",
	Long: `Perintah ini akan melakukan restore multiple databases dari direktori backup files.

Fitur:
- Scan direktori untuk mencari semua file backup
- Ekstrak nama database dari filename menggunakan pattern
- Jika ada multiple files untuk database yang sama, gunakan yang terbaru
- Restore semua databases yang ditemukan
- Support encryption dan compression
- Support safety backup sebelum restore

Pattern filename yang didukung:
  {database}_{year}{month}{day}_{hour}{minute}{second}_{hostname}
  
  Contoh:
  - mydb_20251110_143025_localhost.sql.gz.enc
  - testdb_20251110_143025_dbserver.sql.zst

Contoh penggunaan:
  # Restore dari direktori backup
  sfdbtools restore multi --source /media/ArchiveDB/2025/11/10 --profile prod
  
  # Restore dengan skip pre-backup
  sfdbtools restore multi -s /backup/today --profile prod --skip-backup
  
  # Dry run untuk melihat apa yang akan di-restore
  sfdbtools restore multi -s /backup/today --profile prod --dry-run
  
  # Force restore (continue on errors)
  sfdbtools restore multi -s /backup/today --profile prod --force`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Check dependencies
		if types.Deps == nil {
			return fmt.Errorf("dependencies tidak tersedia")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check dependencies
		if types.Deps == nil {
			return fmt.Errorf("dependencies tidak tersedia")
		}

		logger := types.Deps.Logger
		logger.Info("Memulai proses restore multi database")

		// Parse restore options dari flags
		restoreOpts := types.RestoreOptions{
			Mode: "multi",
		}

		// Parse flag values
		restoreOpts.SourceFile, _ = cmd.Flags().GetString("source")
		restoreOpts.TargetProfile, _ = cmd.Flags().GetString("profile")
		restoreOpts.TargetProfileKey, _ = cmd.Flags().GetString("profile-key")
		restoreOpts.EncryptionKey, _ = cmd.Flags().GetString("encryption-key")
		restoreOpts.Force, _ = cmd.Flags().GetBool("force")
		restoreOpts.DryRun, _ = cmd.Flags().GetBool("dry-run")
		restoreOpts.ShowOptions, _ = cmd.Flags().GetBool("show-options")
		restoreOpts.SkipBackup, _ = cmd.Flags().GetBool("skip-backup")
		restoreOpts.DropTarget, _ = cmd.Flags().GetBool("drop-target")

		// Create restore service
		svc := restore.NewRestoreService(logger, types.Deps.Config, &restoreOpts)

		// Setup context
		ctx := context.Background()

		// Setup restore entry config
		restoreConfig := types.RestoreEntryConfig{
			HeaderTitle: "Restore Multiple Databases",
			ShowOptions: restoreOpts.ShowOptions,
			RestoreMode: "multi",
			SuccessMsg:  "✓ Restore multiple databases selesai",
			LogPrefix:   "[RESTORE-MULTI]",
		}

		// Execute restore command
		if err := svc.ExecuteRestoreCommand(ctx, restoreConfig); err != nil {
			logger.Errorf("Restore multi gagal: %v", err)
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	// Default options untuk restore multi
	defaultOpts := types.RestoreOptions{
		Mode:        "multi",
		Force:       false,
		DryRun:      false,
		ShowOptions: false,
		SkipBackup:  false,
		DropTarget:  false,
	}
	flags.AddRestoreMultiFlags(CmdRestoreMulti, &defaultOpts)
}
