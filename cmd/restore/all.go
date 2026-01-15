// File : cmd/cmd_restore/cmd_restore_all.go
// Deskripsi : Command untuk restore all databases dengan streaming filtering
// Author : Hadiyatna Muflihun
// Tanggal : 18 Desember 2025
// Last Modified : 15 Januari 2026
package restorecmd

import (
	"sfdbtools/internal/app/restore"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/cli/runner"

	"github.com/spf13/cobra"
)

// CmdRestoreAll adalah command untuk restore all databases
var CmdRestoreAll = &cobra.Command{
	Use:   "all",
	Short: "Restore seluruh database instance (Full Instance Restore)",
	Long: `Mengembalikan (Restore) banyak database sekaligus dari file dump besar (misal hasil mariadb-dump atau mysqldump --all-databases).

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
		runner.Run(cmd, func() error {
			_ = restore.ExecuteRestoreAllCommand(cmd, appdeps.Deps)
			return nil
		})
	},
}

func init() {
	flags.AddRestoreAllAllFlags(CmdRestoreAll)
}
