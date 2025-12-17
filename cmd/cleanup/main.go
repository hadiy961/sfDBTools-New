package cleanupcmd

import "github.com/spf13/cobra"

// CmdCleanupMain adalah perintah induk (parent command) untuk operasi pembersihan file backup
var CmdCleanupMain = &cobra.Command{
	Use:     "cleanup",
	Aliases: []string{"clean"},
	Short:   "Bersihkan file backup lama sesuai kebijakan retensi",
	Long: `Perintah untuk membersihkan file-file backup lama di direktori output.
Gunakan sub-perintah untuk menjalankan pembersihan sebenarnya, melihat pratinjau (dry-run),
atau membersihkan berdasarkan pola tertentu.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	CmdCleanupMain.AddCommand(CmdCleanupRun)
}
