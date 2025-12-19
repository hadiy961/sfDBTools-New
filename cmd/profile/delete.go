package profilecmd

import (
	"sfDBTools/internal/profile"
	"sfDBTools/internal/types"
	"sfDBTools/internal/flags"

	"github.com/spf13/cobra"
)

var CmdProfileDelete = &cobra.Command{
	Use:   "delete",
	Short: "Menghapus file profil koneksi database",
	Long: `Menghapus file profil konfigurasi database dari sistem.

Secara default, command ini akan meminta konfirmasi pengguna sebelum melakukan penghapusan untuk mencegah ketidaksengajaan.
Gunakan flag --force untuk melewati konfirmasi (berguna untuk scripting).`,
	Example: `  # 1. Hapus profil dengan konfirmasi (Interaktif memilih jika tidak ada flag)
  sfdbtools profile delete

  # 2. Hapus profil spesifik dengan nama
  sfdbtools profile delete --profile "dev-db"

  # 3. Hapus profil tanpa konfirmasi (Force)
  sfdbtools profile delete --profile "temp-db" --force

  # 4. Hapus profil dari direktori khusus
  sfdbtools profile delete --profile "local-conf" --output-dir "./configs"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return profile.ExecuteProfile(cmd, types.Deps, "delete")
	},
}

func init() {
	flags.ProfileDelete(CmdProfileDelete)
}
