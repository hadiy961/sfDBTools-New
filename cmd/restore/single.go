// File : cmd/cmd_restore/cmd_restore_single.go
// Deskripsi : Command untuk restore single database
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-16

package restorecmd

import (
	"fmt"
	"sfDBTools/internal/restore"
	"sfDBTools/internal/types"

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
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if err := restore.ExecuteRestoreSingleCommand(cmd, types.Deps); err != nil {
			types.Deps.Logger.Error("restore single gagal: " + err.Error())
		}
	},
}

func init() {
	// Profile flags
	CmdRestoreSingle.Flags().StringP("profile", "p", "", "Profile database target untuk restore (ENV: SFDB_TARGET_PROFILE)")
	CmdRestoreSingle.Flags().StringP("profile-key", "k", "", "Kunci enkripsi profile database target (ENV: SFDB_TARGET_PROFILE_KEY)")

	// Encryption flag
	CmdRestoreSingle.Flags().StringP("encryption-key", "K", "", "Kunci enkripsi untuk decrypt file backup (ENV: SFDB_BACKUP_ENCRYPTION_KEY)")

	// Restore options
	CmdRestoreSingle.Flags().Bool("drop-target", true, "Drop target database sebelum restore")
	CmdRestoreSingle.Flags().Bool("skip-backup", false, "Skip backup database target sebelum restore")
	CmdRestoreSingle.Flags().StringP("backup-dir", "b", "", "Direktori output untuk backup pre-restore (default: dari config)")

	// File and database
	CmdRestoreSingle.Flags().StringP("file", "f", "", "Lokasi file backup yang akan di-restore")
	CmdRestoreSingle.Flags().StringP("target-db", "d", "", "Database target untuk restore")

	// User grants
	CmdRestoreSingle.Flags().StringP("grants-file", "g", "", "Lokasi file user grants untuk di-restore (optional)")
	CmdRestoreSingle.Flags().Bool("skip-grants", false, "Skip restore user grants (tidak restore grants sama sekali)")

	// Dry-run mode
	CmdRestoreSingle.Flags().Bool("dry-run", false, "Dry-run mode: validasi file tanpa restore")

	// Ticket (wajib)
	CmdRestoreSingle.Flags().StringP("ticket", "t", "", "Ticket number untuk restore request (wajib)")
}
