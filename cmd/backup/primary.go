// File : cmd/backup/primary.go
// Deskripsi : Command untuk backup database primary
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2026-01-05
package backupcmd

import (
	defaultVal "sfdbtools/internal/cli/defaults"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/pkg/consts"

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

  # 2. Backup primary ke direktori khusus
  sfdbtools db-backup primary --backup-dir "/backups/prod"`,
	Run: func(cmd *cobra.Command, args []string) {
		runBackupCommand(cmd, func() (string, error) {
			return consts.ModePrimary, nil
		})
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions(consts.ModePrimary)
	flags.AddBackupFlgs(CmdBackupPrimary, &defaultOpts, consts.ModePrimary)
}
