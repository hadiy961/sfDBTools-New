// File : cmd/backup/secondary.go
// Deskripsi : Command untuk backup database secondary
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2026-01-23
package backupcmd

import (
	"sfdbtools/internal/app/backup"
	defaultVal "sfdbtools/internal/cli/defaults"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/cli/runner"
	"sfdbtools/internal/shared/consts"

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
  sfdbtools db-backup secondary --backup-dir "/tmp/dev_backups"`,
	Run: func(cmd *cobra.Command, args []string) {
		runner.Run(cmd, func() error {
			return backup.ExecuteBackup(cmd, appdeps.Deps, consts.ModeSecondary)
		})
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions(consts.ModeSecondary)
	flags.AddBackupFlgs(CmdBackupSecondary, &defaultOpts, consts.ModeSecondary)
}
