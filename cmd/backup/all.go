// File : cmd/cmd_backup/cmd_backup_all.go
// Deskripsi : Command untuk backup all databases dengan exclude filters
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-11
// Last Modified : 2025-12-11

package backupcmd

import (
	"fmt"
	"sfDBTools/internal/backup"
	defaultVal "sfDBTools/internal/defaultval"
	"sfDBTools/internal/flags"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"

	"github.com/spf13/cobra"
)

// CmdBackupAll adalah perintah untuk melakukan backup semua database dengan exclude filters
var CmdBackupAll = &cobra.Command{
	Use:   "all",
	Short: "Backup seluruh database instance (Full Instance Backup)",
	Long: `Melakukan backup terhadap SEMUA database yang ada di server dalam satu operasi.

Command ini sangat berguna untuk full server backup. Hasil backup biasanya digabungkan menjadi satu file SQL (tergantung konfigurasi).
Anda dapat mengecualikan database tertentu (misalnya schema sistem MySQL) menggunakan filter exclude.

Fitur:
  - Backup seluruh instance.
  - Filter exclude untuk mengabaikan database sistem atau database tertentu.
  - Dukungan kompresi output (gzip/bzip2/dll tergantung implementasi).`,
	Example: `  # 1. Backup semua database (Default)
  sfdbtools db-backup all

  # 2. Backup semua kecuali database sistem (mysql, information_schema, performance_schema, sys)
  sfdbtools db-backup all --exclude-system

  # 3. Backup ke direktori tertentu dengan kompresi
  sfdbtools db-backup all --output-dir "/backup/daily" --compress

  # 4. Backup dengan mengecualikan list database tertentu
  sfdbtools db-backup all --exclude-db "test_db,temp_db"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		// Backup all menggunakan mode 'all'
		if err := backup.ExecuteBackup(cmd, types.Deps, consts.ModeAll); err != nil {
			// Error has been logged by ExecuteBackup
			return
		}
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions(consts.ModeAll)
	flags.AddBackupAllFlags(CmdBackupAll, &defaultOpts)
}
