// File : cmd/cmd_restore/cmd_restore_single.go
// Deskripsi : Command untuk restore single database
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 5 Januari 2026
package restorecmd

import (
	"fmt"
	"sfdbtools/internal/app/restore"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"

	"github.com/spf13/cobra"
)

// CmdRestoreSingle adalah command untuk restore single database
var CmdRestoreSingle = &cobra.Command{
	Use:   "single",
	Short: "Restore satu database spesifik",
	Long: `Mengembalikan (Restore) satu database tunggal dari file dump.

Command ini menangani berbagai format input secara transparan:
  - File SQL teks biasa (.sql)
  - File terkompresi (.gz, .bz2, dll)
  - File terenkripsi (memerlukan --encryption-key)

Jika nama database target tidak ditentukan, sistem akan mencoba menebak dari nama file atau meminta input pengguna.`,
	Example: `  # 1. Restore database dari file SQL standar
  sfdbtools db-restore single --file "backup.sql" --ticket "TICKET-123"

  # 2. Restore ke nama database target yang berbeda (Rename saat restore)
  sfdbtools db-restore single --file "prod_db.sql" --target-db "dev_db_copy" --ticket "TICKET-123"

  # 3. Restore file terenkripsi
  sfdbtools db-restore single \
    --file "secure_backup.sql.enc" \
    --encryption-key "my-key" \
    --ticket "TICKET-123"

  # 4. Restore tanpa melakukan backup pencegahan (lebih cepat tapi berisiko)
  sfdbtools db-restore single --file "backup.sql" --skip-backup --ticket "TICKET-123"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if appdeps.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if err := restore.ExecuteRestoreSingleCommand(cmd, appdeps.Deps); err != nil {
			appdeps.Deps.Logger.Error("restore single gagal: " + err.Error())
		}
	},
}

func init() {
	flags.AddRestoreSingleFlags(CmdRestoreSingle)
}
