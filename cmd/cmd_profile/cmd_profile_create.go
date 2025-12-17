package cmdprofile

import (
	"sfDBTools/internal/profile"
	"sfDBTools/internal/types"
	"sfDBTools/internal/flags"

	"github.com/spf13/cobra"
)

var CmdProfileCreate = &cobra.Command{
	Use:   "create",
	Short: "Buat profil koneksi database baru",
	Long: `Membuat profil koneksi database baru dan menyimpannya ke konfigurasi lokal.
Anda dapat memberikan detail melalui flag (mis. --profil/-n untuk nama profil, --host/-H, --port/-P, --user/-U).
Tersedia juga opsi seperti --output-dir/-o untuk lokasi penyimpanan, --key/-k untuk kunci enkripsi, dan --interactive/-i untuk mode interaktif.
Jika flag tidak lengkap, proses dapat meminta input secara interaktif.`,
	Example: `
	# Buat profil non-interaktif (gunakan flag --profil/-n untuk nama profil)
	sfdbtools profile create --profil myprofile --host 127.0.0.1 --port 3306 --user root

	# Buat profil interaktif (akan menanyakan nilai yang belum diisi)
	sfdbtools profile create --interactive
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := profile.ExecuteProfile(cmd, types.Deps, "create"); err != nil {
			types.Deps.Logger.Error("profile create gagal: " + err.Error())
		}
	},
}

func init() {
	flags.ProfileCreate(CmdProfileCreate)
}
