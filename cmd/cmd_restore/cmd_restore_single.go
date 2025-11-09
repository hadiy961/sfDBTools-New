package cmdrestore

// File : cmd/cmd_restore/cmd_restore_single.go
// Deskripsi : Command untuk restore database dari satu file backup terpisah
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

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

// CmdRestoreSingle adalah command untuk restore database dari satu file backup terpisah
var CmdRestoreSingle = &cobra.Command{
	Use:   "single",
	Short: "Restore database dari satu file backup terpisah",
	Long: `Perintah ini akan melakukan restore database dari satu file backup terpisah.
	
File backup akan di-decrypt dan di-decompress secara otomatis.
Checksum akan diverifikasi sebelum restore untuk memastikan integritas data.

Contoh penggunaan:
  sfdbtools restore single --source backup/mydb_20251105.sql.gz.enc --target-db mydb_restored
  sfdbtools restore single -s backup/mydb.sql.gz.enc -d mydb -p /path/to/target/profile
  sfdbtools restore single --source backup/mydb.sql.gz.enc --verify-checksum --show-options`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		logger := types.Deps.Logger
		logger.Info("Memulai proses restore single database")

		// Parse restore options dari flags (flags sudah di-define di init())
		restoreOpts := types.RestoreOptions{
			Mode: "single",
		}

		// Parse flag values
		restoreOpts.SourceFile, _ = cmd.Flags().GetString("source")
		restoreOpts.TargetProfile, _ = cmd.Flags().GetString("profile")
		restoreOpts.TargetProfileKey, _ = cmd.Flags().GetString("profile-key")
		restoreOpts.EncryptionKey, _ = cmd.Flags().GetString("encryption-key")
		restoreOpts.TargetDB, _ = cmd.Flags().GetString("target-db")
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
			HeaderTitle: "Database Restore - Single",
			ShowOptions: restoreOpts.ShowOptions,
			RestoreMode: "single",
			SuccessMsg:  "Proses restore single database selesai.",
			LogPrefix:   "[Restore Single]",
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
			logger.Error("restore single gagal: " + err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	// Default options untuk restore single
	defaultOpts := types.RestoreOptions{
		Mode:           "single",
		VerifyChecksum: true,
		Force:          false,
		DryRun:         false,
		ShowOptions:    false,
	}
	flags.AddRestoreSingleFlags(CmdRestoreSingle, &defaultOpts)
}
