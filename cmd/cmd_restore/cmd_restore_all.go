// File : cmd/cmd_restore/cmd_restore_all.go
// Deskripsi : Command untuk restore semua database dari file combined backup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package cmdrestore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sfDBTools/internal/restore"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/flags"
	"syscall"

	"github.com/spf13/cobra"
)

// CmdRestoreAll adalah command untuk restore semua database dari file combined backup
var CmdRestoreAll = &cobra.Command{
	Use:   "all",
	Short: "Restore semua database dari file combined backup",
	Long: `Perintah ini akan melakukan restore semua database dari file combined backup.
	
File combined backup berisi semua database yang di-backup sekaligus.
File akan di-decrypt dan di-decompress secara otomatis.
Checksum akan diverifikasi sebelum restore untuk memastikan integritas data.

Contoh penggunaan:
  sfdbtools restore all --source backup/combined_20251105.sql.gz.enc
  sfdbtools restore all -s backup/all_databases.sql.gz.enc -p /path/to/target/profile
  sfdbtools restore all --source backup/combined.sql.gz.enc --verify-checksum --show-options`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		logger := types.Deps.Logger
		logger.Info("Memulai proses restore all databases")

		// Parse restore options dari flags
		restoreOpts := types.RestoreOptions{
			Mode: "all",
		}

		// Parse flag values
		restoreOpts.SourceFile, _ = cmd.Flags().GetString("source")
		restoreOpts.TargetProfile, _ = cmd.Flags().GetString("profile")
		restoreOpts.TargetProfileKey, _ = cmd.Flags().GetString("profile-key")
		restoreOpts.EncryptionKey, _ = cmd.Flags().GetString("encryption-key")
		restoreOpts.VerifyChecksum, _ = cmd.Flags().GetBool("verify-checksum")
		restoreOpts.Force, _ = cmd.Flags().GetBool("force")
		restoreOpts.DryRun, _ = cmd.Flags().GetBool("dry-run")
		restoreOpts.ShowOptions, _ = cmd.Flags().GetBool("show-options")

		// Inisialisasi restore service
		svc := restore.NewRestoreService(logger, types.Deps.Config, &restoreOpts)

		// Setup context dengan cancellation untuk graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Set cancel function ke service untuk graceful shutdown
		svc.SetCancelFunc(cancel)

		// Setup signal handler untuk CTRL+C (SIGINT) dan SIGTERM
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Goroutine untuk menangani signal
		go func() {
			sig := <-sigChan
			logger.Warnf("Menerima signal %v, menghentikan restore...", sig)
			svc.HandleShutdown()
			cancel()
		}()

		// RestoreEntryConfig menyimpan konfigurasi untuk proses restore
		restoreConfig := types.RestoreEntryConfig{
			HeaderTitle: "Database Restore - All (Combined)",
			ShowOptions: restoreOpts.ShowOptions,
			RestoreMode: "all",
			SuccessMsg:  "Proses restore all databases selesai.",
			LogPrefix:   "[Restore All]",
		}

		if err := svc.ExecuteRestoreCommand(ctx, restoreConfig); err != nil {
			if errors.Is(err, types.ErrUserCancelled) {
				logger.Warn("Proses dibatalkan oleh pengguna.")
				return
			}
			if errors.Is(err, context.Canceled) {
				logger.Warn("Proses restore dibatalkan.")
				return
			}
			logger.Error("restore all gagal: " + err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	// Default options untuk restore all
	defaultOpts := types.RestoreOptions{
		Mode:           "all",
		VerifyChecksum: true,
		Force:          false,
		DryRun:         false,
		ShowOptions:    false,
	}
	flags.AddRestoreAllFlags(CmdRestoreAll, &defaultOpts)
}
