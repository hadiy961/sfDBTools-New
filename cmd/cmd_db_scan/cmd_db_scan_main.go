package cmddbscan

import "github.com/spf13/cobra"

// CmdDBScanMain adalah perintah induk (parent command) untuk semua perintah 'db-scan'.
var CmdDBScanMain = &cobra.Command{
	Use:     "db-scan",
	Aliases: []string{"dbscan"},
	Short:   "Database scanning tools (all, filter, dll)",
	Long: `Perintah 'dbscan' digunakan untuk melakukan scanning database.
Tersedia beberapa sub-perintah seperti all dan filter. Gunakan 'dbscan <sub-command> --help' untuk informasi lebih lanjut.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Tambahkan semua sub-command (Perlu diinisialisasi di file masing-masing)
	CmdDBScanMain.AddCommand(CmdScanAllDB)
	CmdDBScanMain.AddCommand(CmdScanFilter)
	// CmdProfileMain.AddCommand(CmdProfileShow)
	// CmdProfileMain.AddCommand(CmdProfileDelete)
	// CmdProfileMain.AddCommand(CmdProfileEdit)
}
