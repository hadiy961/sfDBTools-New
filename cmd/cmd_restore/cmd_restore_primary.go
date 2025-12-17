// File : cmd/cmd_restore/cmd_restore_primary.go
// Deskripsi : Command untuk restore primary database
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package cmdrestore

import (
	"fmt"
	"sfDBTools/internal/restore"
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
)

// CmdRestorePrimary adalah command untuk restore primary database dengan companion
var CmdRestorePrimary = &cobra.Command{
	Use:   "primary",
	Short: "Restore database primary beserta companion database (dmart)",
	Long: `Command ini akan restore database primary beserta companion database (_dmart).
	
Fitur:
- Restore database primary ke server target
- Auto-detect atau manual select file companion database (_dmart)
- Konfirmasi jika database belum ada
- Setup user aplikasi dengan password dari ENV_PASSWORD_APP
- Backup pre-restore (optional)
- Restore user grants (optional)
	
Hanya database primary dan primary_dmart yang akan di-restore.`,
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

	// Ticket (wajib)
	CmdRestorePrimary.Flags().StringP("ticket", "t", "", "Ticket number untuk restore request (wajib)")
}
