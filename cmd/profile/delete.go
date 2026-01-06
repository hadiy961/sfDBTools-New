package profilecmd

import (
	"sfdbtools/internal/app/profile"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/pkg/consts"

	"github.com/spf13/cobra"
)

var CmdProfileDelete = &cobra.Command{
	Use:   "delete",
	Short: "Menghapus file profil koneksi database",
	Long: `Menghapus file profil konfigurasi database dari sistem.

Secara default, command ini akan meminta konfirmasi pengguna sebelum melakukan penghapusan untuk mencegah ketidaksengajaan.
Gunakan flag --force untuk melewati konfirmasi (berguna untuk scripting).

` + consts.ProfileCLIModeNonInteractiveHeader + `
	- Wajib isi --profile dan --force (tidak ada prompt).`,
	Example: `  # 1) Interaktif (pilih profil jika tidak ada flag)
	sfdbtools profile delete

	# 2) Hapus 1 profil (tetap minta konfirmasi)
	sfdbtools profile delete --profile "dev-db"

	# 3) Non-interaktif (automation/pipeline)
	sfdbtools profile delete --quiet --force --profile "temp-db"

	# 4) Hapus banyak profil (non-interaktif)
	sfdbtools profile delete --quiet --force --profile "a" --profile "b"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return profile.ExecuteProfile(cmd, appdeps.Deps, "delete")
	},
}

func init() {
	flags.ProfileDelete(CmdProfileDelete)
}
