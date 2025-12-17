package cmdprofile

import (
	"sfDBTools/internal/profile"
	"sfDBTools/internal/types"
	"sfDBTools/internal/flags"

	"github.com/spf13/cobra"
)

var CmdProfileEdit = &cobra.Command{
	Use:   "edit",
	Short: "Edit profil koneksi database yang sudah ada",
	Long: `Mengubah nilai pada profil koneksi yang sudah tersimpan.
Anda dapat menjalankan secara non-interaktif dengan flag yang menunjuk field yang ingin diubah, atau menjalankan mode interaktif untuk mengedit nilai langkah demi langkah.
Pastikan profil ditentukan melalui --file/-f agar profil yang tepat dapat ditemukan; untuk mengganti nama gunakan --new-name/-N.`,
	Example: `
	# Edit host pada profil dengan file/nama myprofile
	sfdbtools profile edit --file myprofile --host 192.168.1.10

	# Edit interaktif (akan menanyakan field yang ingin diubah)
	sfdbtools profile edit --file myprofile --interactive
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := profile.ExecuteProfile(cmd, types.Deps, "edit"); err != nil {
			types.Deps.Logger.Error("profile edit gagal: " + err.Error())
		}
	},
}

func init() {
	flags.ProfileEdit(CmdProfileEdit)
}
