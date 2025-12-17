package cmdprofile

import (
	"sfDBTools/internal/profile"
	"sfDBTools/internal/types"
	"sfDBTools/internal/flags"

	"github.com/spf13/cobra"
)

var CmdProfileShow = &cobra.Command{
	Use:   "show",
	Short: "Tampilkan detail profil koneksi",
	Long: `Menampilkan informasi lengkap dari profil koneksi yang tersimpan, seperti host, port, user, database, dan status enkripsi password.
Gunakan flag --file/-f untuk menampilkan profil tertentu. Jika tidak memberikan --file, perintah dapat meminta pilihan secara interaktif (jika didukung).`,
	Example: `
	# Tampilkan detail profil bernama myprofile
	sfdbtools profile show --file myprofile

	# Tampilkan profil dan perlihatkan password secara jelas (jika diizinkan)
	sfdbtools profile show --file myprofile --reveal-password
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return profile.ExecuteProfile(cmd, types.Deps, "show")
	},
}

func init() {
	flags.ProfileShow(CmdProfileShow)
}
