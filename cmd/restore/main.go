// File : cmd/cmd_restore/cmd_restore_main.go
// Deskripsi : Root command untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-16

package restorecmd

import (
	"github.com/spf13/cobra"
)

// CmdRestore adalah root command untuk restore operations
var CmdRestore = &cobra.Command{
	Use:     "db-restore",
	Aliases: []string{"restore", "dbrestore", "import"},
	Short:   "Suite lengkap untuk pemulihan (Restore) database",
	Long: `Kumpulan alat untuk mengembalikan data (Restore) ke server database dari file dump.

Mendukung berbagai skenario pemulihan:
  - Restore Seluruh Instance (all) - Menggunakan streaming filtering untuk efisiensi RAM.
  - Restore Database Tunggal (single) - Cepat dan fleksibel (bisa ganti nama database).
  - Restore Paket Primary (primary) - Memulihkan database utama dan pendampingnya (dmart).

Fitur Keamanan:
  - Wajib Ticket ID: Setiap operasi restore harus menyertakan Ticket ID untuk logging audit.
  - Pre-Restore Backup: Secara default mencadangkan database yang ada sebelum ditimpa.
  - Dekripsi & Dekompresi otomatis.`,
	Example: `  # Lihat bantuan untuk metode restore tertentu
  sfdbtools db-restore all --help
  sfdbtools db-restore single --help
  sfdbtools db-restore primary --help`,
}

func init() {
	// Register subcommands
	CmdRestore.AddCommand(CmdRestoreSingle)
	CmdRestore.AddCommand(CmdRestorePrimary)
	CmdRestore.AddCommand(CmdRestoreAll)
	CmdRestore.AddCommand(CmdRestoreSelection)
}
