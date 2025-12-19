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
	Short: "Restore seluruh database instance (Full Instance Restore)",
	Long: `Mengembalikan (Restore) banyak database sekaligus dari file dump besar (misal hasil mysqldump --all-databases).

Command ini menggunakan teknik "Streaming Filtering" sehingga TIDAK meload seluruh file dump ke RAM,
melainkan memprosesnya baris per baris. Sangat efisien untuk file dump berukuran besar (GB/TB).

Fitur Utama:
  - Streaming Restore: Hemat memori.
  - Drop Target: Opsi untuk menghapus bersih semua database non-sistem sebelum restore.
  - Safety Backup: Melakukan backup otomatis terhadap database yang ada sebelum ditimpa (kecuali --skip-backup).
  - Dry-Run: Simulasi proses restore tanpa melakukan perubahan apa pun.

PERINGATAN: Operasi ini bersifat DESTRUKTIF massal. Pastikan Anda memiliki backup yang valid sebelum menjalankannya!`,
	Example: `  # 1. Restore standar dari file dump besar (memerlukan Ticket ID)
  sfdbtools db-restore all --file "full_backup.sql" --ticket "TICKET-123"

  # 2. Restore dengan menghapus semua database lama terlebih dahulu
  sfdbtools db-restore all --file "full_backup.sql" --ticket "TICKET-123" --drop-target

  # 3. Restore file terkompresi & terenkripsi
  sfdbtools db-restore all --file "full_backup.sql.gz.enc" --encryption-key "my-secret" --ticket "TICKET-123"

  # 4. Simulasi restore (Dry-Run) untuk mengecek isi file dump
  sfdbtools db-restore all --file "full_backup.sql" --dry-run`,
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
	
	// Drop Target
	CmdRestoreAll.Flags().Bool("drop-target", false, "Drop semua database non-sistem sebelum restore")
}
