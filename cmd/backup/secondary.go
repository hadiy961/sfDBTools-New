package backupcmd

import (
	"fmt"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/types"
	defaultVal "sfDBTools/internal/defaultval"
	"sfDBTools/internal/flags"

	"github.com/spf13/cobra"
)

// CmdBackupSecondary adalah perintah untuk melakukan backup database secondary.
var CmdBackupSecondary = &cobra.Command{
	Use:   "secondary",
	Short: "Backup kelompok database 'Secondary' (Smart Filter)",
	Long: `Melakukan backup otomatis HANYA untuk database yang memiliki suffix '_secondary'.

Kebalikan dari command 'primary', perintah ini khusus mencari database development/staging/testing
yang mengikuti konvensi penamaan berakhiran '_secondary'.

Contoh:
  - 'app_db'          -> Primary (Akan diabaikan)
  - 'app_db_secondary'-> Secondary (Akan dibackup)`,
	Example: `  # 1. Backup semua database secondary
  sfdbtools db-backup secondary

  # 2. Backup secondary ke lokasi lain
  sfdbtools db-backup secondary --output-dir "/tmp/dev_backups"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if err := backup.ExecuteBackup(cmd, types.Deps, "secondary"); err != nil {
			// Error has been logged by ExecuteBackup
			return
		}
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions("secondary")
	flags.AddBackupFlgs(CmdBackupSecondary, &defaultOpts, "secondary")
}
