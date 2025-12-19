// File : cmd/cmd_restore/cmd_restore_primary.go
// Deskripsi : Command untuk restore primary database
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package restorecmd

import (
	"fmt"
	"sfDBTools/internal/restore"
	"sfDBTools/internal/types"

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
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if err := restore.ExecuteRestorePrimaryCommand(cmd, types.Deps); err != nil {
			types.Deps.Logger.Error("restore primary gagal: " + err.Error())
		}
	},
}

func init() {
	// Profile flags
	CmdRestorePrimary.Flags().StringP("profile", "p", "", "Profile database target untuk restore (ENV: SFDB_TARGET_PROFILE)")
	CmdRestorePrimary.Flags().StringP("profile-key", "k", "", "Kunci enkripsi profile database target (ENV: SFDB_TARGET_PROFILE_KEY)")

	// Encryption flag
	CmdRestorePrimary.Flags().StringP("encryption-key", "K", "", "Kunci enkripsi untuk decrypt file backup (ENV: SFDB_BACKUP_ENCRYPTION_KEY)")

	// Restore options
	CmdRestorePrimary.Flags().Bool("drop-target", true, "Drop target database sebelum restore")
	CmdRestorePrimary.Flags().Bool("skip-backup", false, "Skip backup database target sebelum restore")
	CmdRestorePrimary.Flags().StringP("backup-dir", "b", "", "Direktori output untuk backup pre-restore (default: dari config)")

	// File and database
	CmdRestorePrimary.Flags().StringP("file", "f", "", "Lokasi file backup primary yang akan di-restore")
	CmdRestorePrimary.Flags().StringP("companion-file", "c", "", "Lokasi file backup companion (_dmart) - optional, auto-detect jika kosong")
	CmdRestorePrimary.Flags().StringP("target-db", "d", "", "Database primary target untuk restore")

	// Companion options
	CmdRestorePrimary.Flags().Bool("include-dmart", true, "Include restore companion database (_dmart)")
	CmdRestorePrimary.Flags().Bool("auto-detect-dmart", true, "Auto-detect file companion database (_dmart)")
	CmdRestorePrimary.Flags().Bool("skip-confirm", false, "Skip konfirmasi jika database belum ada")

	// User grants
	CmdRestorePrimary.Flags().StringP("grants-file", "g", "", "Lokasi file user grants untuk di-restore (optional)")
	CmdRestorePrimary.Flags().Bool("skip-grants", false, "Skip restore user grants (tidak restore grants sama sekali)")

	// Dry-run mode
	CmdRestorePrimary.Flags().Bool("dry-run", false, "Dry-run mode: validasi file tanpa restore")

	// Ticket (wajib)
	CmdRestorePrimary.Flags().StringP("ticket", "t", "", "Ticket number untuk restore request (wajib)")
}
