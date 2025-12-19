package profilecmd

import (
	"sfDBTools/internal/profile"
	"sfDBTools/internal/types"
	"sfDBTools/internal/flags"

	"github.com/spf13/cobra"
)

var CmdProfileShow = &cobra.Command{
	Use:   "show",
	Short: "Menampilkan detail informasi profil",
	Long: `Menampilkan isi konfigurasi dari sebuah profil database.

Secara default, informasi sensitif seperti password akan disensor (ditampilkan sebagai bintang/masking).
Gunakan flag --reveal-password untuk menampilkan password dalam bentuk teks biasa (plaintext).
Jika profil terenkripsi, Anda mungkin perlu memasukkan kunci enkripsi profil.`,
	Example: `  # 1. Pilih profil secara interaktif untuk ditampilkan
  sfdbtools profile show

  # 2. Tampilkan profil tertentu
  sfdbtools profile show --profile "dev-db"

  # 3. Tampilkan profil dengan password terlihat (Unmasked)
  sfdbtools profile show --profile "dev-db" --reveal-password

  # 4. Tampilkan profil yang berada di direktori khusus
  sfdbtools profile show --profile "custom-conf" --output-dir "./configs"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return profile.ExecuteProfile(cmd, types.Deps, "show")
	},
}

func init() {
	flags.ProfileShow(CmdProfileShow)
}
