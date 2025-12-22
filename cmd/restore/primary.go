// File : cmd/cmd_restore/cmd_restore_primary.go
// Deskripsi : Command untuk restore primary database
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package restorecmd

import (
	"fmt"
	appdeps "sfDBTools/internal/deps"
	"sfDBTools/internal/flags"
	"sfDBTools/internal/restore"

	"github.com/spf13/cobra"
)

// CmdRestorePrimary adalah command untuk restore primary database dengan companion
var CmdRestorePrimary = &cobra.Command{
	Use:   "primary",
	Short: "Restore paket database Primary & Companion",
	Long: `Mengembalikan (Restore) database utama (Primary) beserta pasangannya (Companion/Dmart).

Seringkali sebuah aplikasi memiliki database transaksional utama dan database reporting/warehouse pendamping (misal: 'app_db' dan 'app_db_dmart').
Command ini otomatis menangani keduanya dalam satu rangkaian proses.

Fitur:
  - Auto-Discovery: Mencoba menemukan file dump companion (_dmart) secara otomatis jika tidak dispesifikasikan.
  - User Grants: Dapat memulihkan hak akses user (Grants) setelah restore data selesai.
  - Environment Password: Setup password user aplikasi secara otomatis via environment variable.`,
	Example: `  # 1. Restore paket primary (auto-detect companion)
  sfdbtools db-restore primary --file "app_db_backup.sql" --ticket "TICKET-123"

  # 2. Restore primary dengan menunjuk file companion secara eksplisit
  sfdbtools db-restore primary \
    --file "app_db.sql" \
    --companion-file "app_db_dmart.sql" \
    --ticket "TICKET-123"

  # 3. Restore primary saja (abaikan dmart)
  sfdbtools db-restore primary --file "app_db.sql" --include-dmart=false --ticket "TICKET-123"

  # 4. Restore dan apply user grants dari file
  sfdbtools db-restore primary --file "app_db.sql" --grants-file "users.sql" --ticket "TICKET-123"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if appdeps.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if err := restore.ExecuteRestorePrimaryCommand(cmd, appdeps.Deps); err != nil {
			appdeps.Deps.Logger.Error("restore primary gagal: " + err.Error())
		}
	},
}

func init() {
	flags.AddRestorePrimaryAllFlags(CmdRestorePrimary)
}
