// File : cmd/backup/all.go
// Deskripsi : Command untuk backup all databases dengan exclude filters
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-11
// Last Modified : 2026-01-05
package backupcmd

import (
	defaultVal "sfdbtools/internal/cli/defaults"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/shared/consts"

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
`,
	Example: `  # 1. Backup semua database (Default)
  sfdbtools db-backup all

	# 2. Backup ke direktori tertentu
	sfdbtools db-backup all --backup-dir "/backup/daily"

	# 3. Backup dengan custom nama file output
	sfdbtools db-backup all --backup-dir "/backup/daily" --filename "all_backup_20251224"`,
	Run: func(cmd *cobra.Command, args []string) {
		runBackupCommand(cmd, func() (string, error) {
			return consts.ModeAll, nil
		})
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions(consts.ModeAll)
	flags.AddBackupAllFlags(CmdBackupAll, &defaultOpts)
}
