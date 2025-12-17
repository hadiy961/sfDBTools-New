package profilecmd

import "github.com/spf13/cobra"

// CmdProfileMain adalah perintah induk (parent command) untuk semua perintah 'profile'.
var CmdProfileMain = &cobra.Command{
	Use:   "profile",
	Short: "Kelola profil koneksi database pengguna",
	Long: `Perintah 'profile' digunakan untuk mengelola profil koneksi database yang tersimpan secara lokal.
Profil menyimpan detail koneksi (mis. host, port, user, database, opsi enkripsi) dan dapat digunakan oleh perintah lain di sfDBTools.
Gunakan 'sfdbtools profile <sub-command> --help' untuk informasi lebih lanjut tentang setiap sub-perintah (create, edit, delete, show).`,
	Example: `
	# Buat profil baru (non-interaktif)
	sfdbtools profile create --profil myprofile --host 127.0.0.1 --user admin

	# Tampilkan profil (menggunakan nama file/profil)
	sfdbtools profile show --file myprofile
`,
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
