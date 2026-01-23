package cleanupcmd

import "github.com/spf13/cobra"

// CmdCleanupMain adalah perintah induk (parent command) untuk operasi pembersihan file backup
var CmdCleanupMain = &cobra.Command{
	Use:     "cleanup",
	Aliases: []string{"clean"},
	Short:   "Bersihkan file backup lama sesuai kebijakan retensi",
	Long: `Perintah untuk membersihkan file-file backup lama di direktori output.
Gunakan sub-perintah untuk menjalankan pembersihan, atau membersihkan berdasarkan pola tertentu.

Gunakan flag --dry-run untuk melihat pratinjau tanpa menghapus file.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Tambahan subcommand lain (pattern, run, dll) bisa ditambahkan di sini di masa depan
}
