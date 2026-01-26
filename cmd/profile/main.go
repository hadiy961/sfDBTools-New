package profilecmd

import "github.com/spf13/cobra"

// CmdProfileMain adalah perintah induk (parent command) untuk semua perintah 'profile'.
var CmdProfileMain = &cobra.Command{
	Use:   "profile",
	Short: "Manajemen profil koneksi database",
	Long: `Kumpulan command untuk mengelola profil koneksi database (Create, Read, Update, Delete).

Profil menyimpan informasi koneksi (Host, Port, User, Password, dll) sehingga Anda tidak perlu
mengetikkannya berulang kali saat menjalankan operasi backup, restore, atau scanning.
Profil disimpan secara lokal dan dapat diamankan dengan enkripsi.`,
	Example: `  # Buat profil baru
  sfdbtools profile create

  # Lihat daftar/detail profil
  sfdbtools profile show

  # Edit profil yang ada
  sfdbtools profile edit --profile "my-db" --host "localhost"

  # Hapus profil
  sfdbtools profile delete --profile "unused-db"`,
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
	CmdProfileMain.AddCommand(CmdProfileClone)
	CmdProfileMain.AddCommand(CmdProfileImport)
}
