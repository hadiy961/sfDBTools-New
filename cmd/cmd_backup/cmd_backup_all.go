// File : cmd/cmd_backup/cmd_backup_all.go
// Deskripsi : Command untuk backup all databases dengan exclude filters
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-11
// Last Modified : 2025-12-11

package cmdbackup

import (
	"fmt"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/types"
	defaultVal "sfDBTools/internal/defaultval"
	"sfDBTools/internal/flags"

	"github.com/spf13/cobra"
)

// CmdDBBackupAll adalah perintah untuk melakukan backup semua database dengan exclude filters
var CmdDBBackupAll = &cobra.Command{
	Use:   "all",
	Short: "Backup semua database dalam satu file dengan exclude filters",
	Long: `Perintah ini akan melakukan backup semua database dalam satu file (combined mode).
Anda dapat menggunakan exclude filters untuk mengecualikan database tertentu dari backup.

Contoh penggunaan:
  sfdbtools backup all                                    # Backup semua database
  sfdbtools backup all --exclude-system                   # Exclude system databases
  sfdbtools backup all --exclude-db test --exclude-db dev # Exclude specific databases
  sfdbtools backup all --exclude-db-file exclude.txt      # Exclude dari file`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		// Backup all menggunakan mode 'all'
		if err := backup.ExecuteBackup(cmd, types.Deps, "all"); err != nil {
			types.Deps.Logger.Error("db-backup all gagal: " + err.Error())
		}
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions("all")
	flags.AddBackupAllFlags(CmdDBBackupAll, &defaultOpts)
}
