package profilecmd

import (
	"sfdbtools/internal/app/profile"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/shared/consts"

	"github.com/spf13/cobra"
)

var CmdProfileImport = &cobra.Command{
	Use:   "import",
	Short: "Import profil database dari XLSX lokal atau Google Spreadsheet" + consts.ProfileCLIAutoInteractiveSuffix,
	Long: `Import profil database secara bulk.

Sumber import:
	- XLSX lokal: gunakan --input <file.xlsx>
	- Google Spreadsheet: gunakan --gsheet <url> dan --gid <gid>

Aturan utama:
	- 1 row = 1 profile = 1 encryption key.
	- Validasi 3 tahap (schema, per-row, pre-save) sebelum menyimpan.
	- Default conn-test ON.
	- Default on-conflict=fail.

Automation:
	- Wajib gunakan --skip-confirm agar tidak ada prompt.
	- Jika --on-conflict=rename, rename otomatis pakai suffix (_2, _3, dst).
`,
	Example: `  # 1) Import dari XLSX lokal (interaktif)
	sfdbtools profile import --input ./profiles.xlsx --sheet Profiles

	# 2) Import dari Google Spreadsheet (public via link)
	sfdbtools profile import --gsheet "https://docs.google.com/spreadsheets/d/<id>/edit?usp=sharing" --gid 0

	# 3) Automation (tanpa prompt)
	sfdbtools profile import --quiet --skip-confirm \
	  --input ./profiles.xlsx \
	  --on-conflict=rename \
	  --continue-on-error`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return profile.ExecuteProfile(cmd, appdeps.Deps, "import")
	},
}

func init() {
	flags.ProfileImport(CmdProfileImport)
}
