package cmdprofile

import "github.com/spf13/cobra"

// CmdProfileMain adalah perintah induk (parent command) untuk semua perintah 'profile'.
var CmdProfileMain = &cobra.Command{
	Use:   "profile",
	Short: "Mengelola profil pengguna (generate, edit, delete, validate, show)",
	Long: `Perintah 'profile' digunakan untuk mengelola file profil pengguna.
Terdapat beberapa sub-perintah seperti create, edit, delete, validate, dan show.
Gunakan 'profile <sub-command> --help' untuk informasi lebih lanjut tentang masing-masing sub-perintah.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Tambahkan semua sub-command (Perlu diinisialisasi di file masing-masing)
	CmdProfileMain.AddCommand(CmdProfileCreate)
	CmdProfileMain.AddCommand(CmdProfileShow)
	CmdProfileMain.AddCommand(CmdProfileDelete)
	CmdProfileMain.AddCommand(CmdProfileEdit)
}
