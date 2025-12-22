package backupcmd

import (
	"fmt"
	"sfDBTools/internal/backup"
	defaultVal "sfDBTools/internal/defaultval"
	appdeps "sfDBTools/internal/deps"
	"sfDBTools/internal/flags"
	"sfDBTools/pkg/consts"

	"github.com/spf13/cobra"
)

// CmdBackupPrimary adalah perintah untuk melakukan backup database primary.
var CmdBackupPrimary = &cobra.Command{
	Use:   "primary",
	Short: "Backup kelompok database 'Primary' (Smart Filter)",
	Long: `Melakukan backup otomatis untuk semua database yang DIANGGAP sebagai database 'Primary'.

Sistem mendeteksi database 'Primary' dengan cara MENGECUALIKAN database yang memiliki suffix '_secondary'.
Ini berguna jika Anda memiliki konvensi penamaan database development/staging dengan akhiran '_secondary'.

Contoh:
  - 'app_db'          -> Primary (Akan dibackup)
  - 'app_db_secondary'-> Secondary (Akan diabaikan)`,
	Example: `  # 1. Backup semua database primary
  sfdbtools db-backup primary

  # 2. Backup primary dengan kompresi
  sfdbtools db-backup primary --compress

  # 3. Backup primary ke direktori khusus
  sfdbtools db-backup primary --output-dir "/backups/prod"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if appdeps.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if err := backup.ExecuteBackup(cmd, appdeps.Deps, consts.ModePrimary); err != nil {
			// Error has been logged by ExecuteBackup
			return
		}
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions(consts.ModePrimary)
	flags.AddBackupFlgs(CmdBackupPrimary, &defaultOpts, consts.ModePrimary)
}
