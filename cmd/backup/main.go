package backupcmd

import "github.com/spf13/cobra"

// CmdBackupMain adalah perintah induk (parent command) untuk semua perintah 'db-backup'.
var CmdBackupMain = &cobra.Command{
	Use:     "db-backup",
	Aliases: []string{"backup", "dbbackup", "dump"},
	Short:   "Suite lengkap untuk backup database",
	Long: `Kumpulan alat untuk melakukan backup database dengan berbagai strategi.

Mendukung berbagai skenario backup:
  - Backup Seluruh Instance (all)
  - Backup Selektif/Bulk (filter)
  - Backup Database Tunggal (single)
  - Backup Berbasis Konvensi (primary/secondary)

Setiap command mendukung opsi standar seperti kompresi, enkripsi (opsional), dan custom output.`,
	Example: `  # Lihat bantuan untuk command spesifik
  sfdbtools db-backup all --help
  sfdbtools db-backup single --help
  sfdbtools db-backup filter --help`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Tambahkan semua sub-command (Perlu diinisialisasi di file masing-masing)
	CmdBackupMain.AddCommand(CmdBackupAll)
	CmdBackupMain.AddCommand(CmdBackupFilter)
	CmdBackupMain.AddCommand(CmdBackupSingle)
	CmdBackupMain.AddCommand(CmdBackupPrimary)
	CmdBackupMain.AddCommand(CmdBackupSecondary)
}
