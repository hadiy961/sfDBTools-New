package profilecmd

import (
	appdeps "sfDBTools/internal/cli/deps"
	"sfDBTools/internal/cli/flags"
	"sfDBTools/internal/profile"
	"sfDBTools/pkg/consts"

	"github.com/spf13/cobra"
)

var CmdProfileShow = &cobra.Command{
	Use:   "show",
	Short: "Menampilkan detail profil" + consts.ProfileCLIAutoInteractiveSuffix,
	Long: `Menampilkan isi konfigurasi dari sebuah profil database.

Default:
	- Tanpa --quiet (TTY): bisa pilih profil dan input key secara interaktif.

` + consts.ProfileCLIModeNonInteractiveHeader + `
	- Wajib isi --profile dan --profile-key ` + consts.ProfileCLINonInteractiveEnvProfileKeyNote + `

Catatan:
	- Password disensor secara default.
	- Gunakan --reveal-password untuk menampilkan password dalam bentuk teks biasa.`,
	Example: `  # 1) Interaktif (pilih profil + input key jika perlu)
	sfdbtools profile show

	# 2) Non-interaktif
	sfdbtools profile show --quiet --profile "dev-db" --profile-key "my-key"

	# 3) Non-interaktif via environment variables
	SFDB_TARGET_PROFILE_KEY='my-key' sfdbtools profile show --quiet --profile "dev-db"

	# 4) Tampilkan password (hati-hati)
	sfdbtools profile show --quiet --profile "dev-db" --profile-key "my-key" --reveal-password`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return profile.ExecuteProfile(cmd, appdeps.Deps, "show")
	},
}

func init() {
	flags.ProfileShow(CmdProfileShow)
}
