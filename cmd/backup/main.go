package backupcmd

import "github.com/spf13/cobra"

// CmdDBBackupMain adalah perintah induk (parent command) untuk semua perintah 'db-backup'.
var CmdDBBackupMain = &cobra.Command{
	Use:     "db-backup",
	Aliases: []string{"backup", "dbbackup", "dump"},
	Short:   "Database backup tools (all, filter, single, dll)",
	Long: `Perintah 'db-backup' digunakan untuk melakukan backup database.
Tersedia beberapa sub-perintah seperti all, filter, dan single. Gunakan 'db-backup <sub-command> --help' untuk informasi lebih lanjut.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Tambahkan semua sub-command (Perlu diinisialisasi di file masing-masing)
	CmdDBBackupMain.AddCommand(CmdDBBackupAll)
	CmdDBBackupMain.AddCommand(CmdDBBackupFilter)
	CmdDBBackupMain.AddCommand(CmdBackupSingle)
	CmdDBBackupMain.AddCommand(CmdBackupPrimary)
	CmdDBBackupMain.AddCommand(CmdBackupSecondary)
}
