// File : cmd/cmd_restore/cmd_restore_all.go
// Deskripsi : Command untuk restore all databases dengan streaming filtering
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-18
// Last Modified : 2025-12-18

package restorecmd

import (
	"fmt"
	"sfDBTools/internal/restore"
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
)

// CmdRestoreAll adalah command untuk restore all databases
var CmdRestoreAll = &cobra.Command{
	Use:   "all",
	Short: "Restore semua database dari file dump dengan streaming filtering",
	Long: `Command ini akan restore semua database dari file backup hasil mysqldump --all-databases.
	
Fitur utama:
  • Streaming processing - tidak load seluruh file ke RAM
  • Filtering realtime - skip database tertentu di tengah proses
  • Skip system databases - otomatis skip mysql, sys, information_schema, performance_schema
  • Dry-run mode - analisis file tanpa restore
  • Safety backup - backup sebelum restore (optional)

PERHATIAN: Operasi ini berisiko karena akan restore/overwrite banyak database sekaligus!`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		if err := restore.ExecuteRestoreAllCommand(cmd, types.Deps); err != nil {
			types.Deps.Logger.Error("restore all gagal: " + err.Error())
		}
	},
}

func init() {
	// Profile flags
	CmdRestoreAll.Flags().StringP("profile", "p", "", "Profile database target untuk restore (ENV: SFDB_TARGET_PROFILE)")
	CmdRestoreAll.Flags().StringP("profile-key", "k", "", "Kunci enkripsi profile database target (ENV: SFDB_TARGET_PROFILE_KEY)")

	// Encryption flag
	CmdRestoreAll.Flags().StringP("encryption-key", "K", "", "Kunci enkripsi untuk decrypt file backup (ENV: SFDB_BACKUP_ENCRYPTION_KEY)")

	// File
	CmdRestoreAll.Flags().StringP("file", "f", "", "Lokasi file backup all-databases yang akan di-restore")

	// Ticket (wajib)
	CmdRestoreAll.Flags().StringP("ticket", "t", "", "Ticket number untuk restore request (wajib)")

	// Safety & Behavior flags
	CmdRestoreAll.Flags().Bool("skip-backup", false, "Skip backup sebelum restore")
	CmdRestoreAll.Flags().StringP("backup-dir", "b", "", "Direktori output untuk backup pre-restore (default: dari config)")
	CmdRestoreAll.Flags().Bool("dry-run", false, "Dry-run mode: analisis file tanpa restore")
	CmdRestoreAll.Flags().Bool("force", false, "Force restore tanpa konfirmasi interaktif")
	CmdRestoreAll.Flags().Bool("continue-on-error", false, "Lanjutkan restore meski ada error (default: stop on error)")

	// Filtering flags
	CmdRestoreAll.Flags().StringSlice("exclude-dbs", []string{}, "Daftar database yang akan di-exclude (dipisah koma)")
	CmdRestoreAll.Flags().Bool("include-system", false, "Include system databases (mysql, sys, dll) - default di-skip untuk keamanan")
}
