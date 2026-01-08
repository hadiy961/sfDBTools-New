package profilecmd

import (
	"sfdbtools/internal/app/profile"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/shared/consts"

	"github.com/spf13/cobra"
)

var CmdProfileCreate = &cobra.Command{
	Use:   "create",
	Short: "Membuat profil koneksi database baru" + consts.ProfileCLIAutoInteractiveSuffix,
	Long: `Membuat profil koneksi database baru untuk digunakan oleh sfdbtools.

Command ini akan menyimpan kredensial dan konfigurasi koneksi ke dalam file profil.
Profil ini nantinya digunakan untuk operasi backup, restore, dan dbscan tanpa perlu memasukkan ulang kredensial.

Mode:
  - Interaktif (default): jika TTY dan tanpa --quiet, akan memandu input (wizard).
  - Non-interaktif (--quiet): wajib isi parameter lewat flag/env (tanpa prompt).

Validasi:
  - sfdbtools akan mencoba koneksi DB sebelum menyimpan profil.
  - Jika koneksi gagal: mode interaktif bisa pilih lanjut/batal, mode --quiet akan gagal (fail-fast).
`,
	Example: `  # 1) Interaktif (wizard)
  sfdbtools profile create

  # 2) Non-interaktif (automation/pipeline)
  sfdbtools profile create --quiet \
    --profile "prod-db" \
    --host "10.0.0.5" \
    --port 3306 \
    --user "admin" \
    --password "s3cr3t"

  # 3) Non-interaktif via environment variables
  SFDB_TARGET_DB_HOST=10.0.0.5 \
  SFDB_TARGET_DB_USER=admin \
  SFDB_TARGET_DB_PASSWORD='s3cr3t' \
  SFDB_TARGET_PROFILE_KEY='my-secret-key-123' \
  sfdbtools profile create --quiet --profile "prod-db"

  # 4) Profil terenkripsi (wajib di mode --quiet)
  sfdbtools profile create --quiet \
    --profile "secure-db" \
    --host "localhost" \
    --user "root" \
    --password "toor" \
    --profile-key "my-secret-key-123"`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := profile.ExecuteProfile(cmd, appdeps.Deps, "create"); err != nil {
			appdeps.Deps.Logger.Error("profile create gagal: " + err.Error())
		}
	},
}

func init() {
	flags.ProfileCreate(CmdProfileCreate)
}
