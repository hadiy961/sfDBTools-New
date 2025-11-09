package cmdbackup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/types"
	defaultVal "sfDBTools/pkg/defaultval"
	"sfDBTools/pkg/flags"
	"sfDBTools/pkg/parsing"
	"syscall"

	"github.com/spf13/cobra"
)

// CmdDBBackupCombined adalah perintah untuk melakukan backup database secara combined
var CmdDBBackupCombined = &cobra.Command{
	Use:   "combined",
	Short: "Backup database secara combined",
	Long:  `Perintah ini akan melakukan backup database dengan metode combined.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		logger := types.Deps.Logger
		logger.Info("Memulai proses backup database secara combined")

		// Parsing opsi memakai parser baru agar konsisten dengan command filter
		parsedOpts, err := parsing.ParsingBackupOptions(cmd, "combined")
		if err != nil {
			logger.Error("gagal parsing opsi: " + err.Error())
			return
		}

		// Inisialisasi service backup
		svc := backup.NewBackupService(logger, types.Deps.Config, &parsedOpts)

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
			logger.Warnf("Menerima signal %v, menghentikan backup...", sig)
			svc.HandleShutdown()
			cancel()
		}()

		// BackupEntryConfig menyimpan konfigurasi untuk proses backup
		backupConfig := types.BackupEntryConfig{
			HeaderTitle: "Database Backup - Combined",
			Force:       parsedOpts.Force,
			SuccessMsg:  "Proses backup database secara combined selesai.",
			LogPrefix:   "[Backup Combined]",
			BackupMode:  "combined",
		}

		if err := svc.ExecuteBackupCommand(ctx, backupConfig); err != nil {
			if errors.Is(err, types.ErrUserCancelled) {
				logger.Warn("Proses dibatalkan oleh pengguna.")
				return
			}
			if errors.Is(err, context.Canceled) {
				logger.Warn("Proses backup dibatalkan.")
				return
			}
			logger.Error("db-backup combined gagal: " + err.Error())
		}

	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions("combined") // Tambahkan flags untuk backup combined
	flags.AddBackupFlags(CmdDBBackupCombined, &defaultOpts)
}
