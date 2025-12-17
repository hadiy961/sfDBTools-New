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
	Short: "Restore satu database dari file backup",
	Long:  `Command ini akan restore satu database dari file backup dengan opsi decrypt dan decompress otomatis.`,
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

	// Ticket (wajib)
	CmdRestoreSingle.Flags().StringP("ticket", "t", "", "Ticket number untuk restore request (wajib)")
}
