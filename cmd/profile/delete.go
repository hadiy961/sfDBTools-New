package profilecmd

import (
	"sfDBTools/internal/profile"
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
)

var CmdProfileDelete = &cobra.Command{
	Use:   "delete",
	Short: "Hapus profil koneksi database",
	Long: `Menghapus profil koneksi yang tersimpan setelah konfirmasi pengguna.
	Secara default perintah akan meminta konfirmasi sebelum menghapus. Gunakan flag --force/-F untuk melewati konfirmasi.
	Perintah ini hanya menghapus metadata profil pada konfigurasi lokal, tidak mengubah data pada server database.`,
	Example: `
		# Hapus profil dengan konfirmasi (gunakan --file/-f untuk menunjuk profil)
		sfdbtools profile delete --file myprofile

		# Hapus profil tanpa konfirmasi (force)
		sfdbtools profile delete --file oldprofile --force
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return profile.ExecuteProfile(cmd, types.Deps, "delete")
	},
}
