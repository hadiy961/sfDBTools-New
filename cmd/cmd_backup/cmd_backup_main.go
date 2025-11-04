package cmdbackup

import "github.com/spf13/cobra"

// CmdDBBackupMain adalah perintah induk (parent command) untuk semua perintah 'db-backup'.
var CmdDBBackupMain = &cobra.Command{
	Use:     "db-backup",
	Aliases: []string{"backup", "dbbackup", "dump"},
	Short:   "Database backup tools (combined, separated, dll)",
	Long: `Perintah 'db-backup' digunakan untuk melakukan backup database.
Tersedia beberapa sub-perintah seperti combined dan separated. Gunakan 'db-backup <sub-command> --help' untuk informasi lebih lanjut.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Tambahkan semua sub-command (Perlu diinisialisasi di file masing-masing)
	CmdDBBackupMain.AddCommand(CmdDBBackupCombined)
	CmdDBBackupMain.AddCommand(CmdDBBackupSeparated)
}
