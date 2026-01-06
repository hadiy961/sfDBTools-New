package profilecmd

import (
	"sfdbtools/internal/app/profile"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/pkg/consts"

	"github.com/spf13/cobra"
)

var CmdProfileEdit = &cobra.Command{
	Use:   "edit",
	Short: "Mengubah profil database yang ada" + consts.ProfileCLIAutoInteractiveSuffix,
	Long: `Mengubah atau memperbarui nilai konfigurasi dalam file profil yang sudah ada.

Mode:
	- Interaktif (default): jika TTY dan tanpa --quiet, akan memilih profil dan field yang diubah.
	- Non-interaktif (--quiet): wajib tentukan profil target lewat --profile dan sertakan --profile-key ` + consts.ProfileCLINonInteractiveEnvProfileKeyNote + `

Command ini juga mendukung rename menggunakan flag --new-name.`,
	Example: `  # 1) Interaktif (pilih profil + field)
	sfdbtools profile edit

	# 2) Non-interaktif (automation/pipeline)
	sfdbtools profile edit --quiet \
		--profile "prod-db" \
		--profile-key "my-key" \
		--host "10.0.0.6" \
		--port 3307

	# 3) Ganti password
	sfdbtools profile edit --quiet --profile "prod-db" --profile-key "my-key" --password "new-password-2025"

	# 4) Rename
	sfdbtools profile edit --quiet --profile "old-name" --profile-key "my-key" --new-name "new-name"

	# 5) Profil terenkripsi (opsional)
	sfdbtools profile edit --quiet --profile "secure-db" --host "1.2.3.4" --profile-key "my-key"

	# 6) Rotasi kunci enkripsi profil
	sfdbtools profile edit --quiet --profile "secure-db" --profile-key "old-key" --new-profile-key "new-key"`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := profile.ExecuteProfile(cmd, appdeps.Deps, "edit"); err != nil {
			appdeps.Deps.Logger.Error("profile edit gagal: " + err.Error())
		}
	},
}

func init() {
	flags.ProfileEdit(CmdProfileEdit)
}
