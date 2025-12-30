// File : cmd/backup/single.go
// Deskripsi : Command untuk backup satu database
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package backupcmd

import (
	defaultVal "sfDBTools/internal/defaultval"
	"sfDBTools/internal/flags"
	"sfDBTools/pkg/consts"

	"github.com/spf13/cobra"
)

// CmdBackupSingle adalah perintah untuk melakukan backup satu database.
var CmdBackupSingle = &cobra.Command{
	Use:   "single",
	Short: "Backup satu database spesifik",
	Long: `Melakukan backup untuk SATU database saja.

Command ini dioptimalkan untuk backup cepat database tunggal.
Jika nama database tidak diberikan via flag, akan muncul menu interaktif untuk memilih satu database.`,
	Example: `  # 1. Pilih satu database secara interaktif
  sfdbtools db-backup single

  # 2. Backup database tertentu
  sfdbtools db-backup single --db "target_db"

  # 3. Backup ke output file spesifik
  sfdbtools db-backup single --db "target_db" --output-file "backup_target_v1.sql"
`,
	Run: func(cmd *cobra.Command, args []string) {
		runBackupCommand(cmd, func() (string, error) {
			return consts.ModeSingle, nil
		})
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions(consts.ModeSingle)
	flags.AddBackupFlgs(CmdBackupSingle, &defaultOpts, consts.ModeSingle)
}
