package cmdbackup

import (
	"errors"
	"fmt"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/types"
	defaultVal "sfDBTools/pkg/defaultval"
	"sfDBTools/pkg/flags"
	"sfDBTools/pkg/parsing"

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

		// BackupEntryConfig menyimpan konfigurasi untuk proses backup
		backupConfig := types.BackupEntryConfig{
			HeaderTitle: "Database Backup - Combined",
			ShowOptions: parsedOpts.ShowOptions,
			SuccessMsg:  "Proses backup database secara combined selesai.",
			LogPrefix:   "[Backup Combined]",
			BackupMode:  "combined",
		}

		if err := svc.ExecuteBackupCommand(backupConfig); err != nil {
			if errors.Is(err, types.ErrUserCancelled) {
				logger.Warn("Proses dibatalkan oleh pengguna.")
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
