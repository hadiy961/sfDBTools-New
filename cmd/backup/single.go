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

  # 4. Backup dengan kompresi
  sfdbtools db-backup single --db "target_db" --compress`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if appdeps.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if err := backup.ExecuteBackup(cmd, appdeps.Deps, consts.ModeSingle); err != nil {
			// Error has been logged by ExecuteBackup
			return
		}
	},
}

func init() {
	defaultOpts := defaultVal.DefaultBackupOptions(consts.ModeSingle)
	flags.AddBackupFlgs(CmdBackupSingle, &defaultOpts, consts.ModeSingle)
}
