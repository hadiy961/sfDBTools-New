package profilecmd

import (
	"sfdbtools/internal/app/profile"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/shared/consts"

	"github.com/spf13/cobra"
)

var CmdProfileClone = &cobra.Command{
	Use:   "clone",
	Short: "Clone profil koneksi database yang sudah ada" + consts.ProfileCLIAutoInteractiveSuffix,
	Long: `Clone profil koneksi database yang sudah ada dengan modifikasi minimal.

Command ini berguna untuk setup replicas atau environment berbeda dengan konfigurasi serupa.
Source profile akan dimuat, lalu Anda bisa mengubah beberapa field (seperti host, port, nama)
sebelum disimpan sebagai profil baru.

Mode:
  - Interaktif (default): wizard dengan pre-fill dari source profile.
  - Non-interaktif (--quiet): clone langsung dengan override dari flags (tanpa prompt).

Validasi:
  - sfdbtools akan mencoba koneksi DB ke profil hasil clone sebelum menyimpan.
  - Jika koneksi gagal: mode interaktif bisa pilih lanjut/batal, mode --quiet akan gagal (fail-fast).
`,
	Example: `  # 1) Interaktif (wizard dengan pre-fill)
  sfdbtools profile clone --source "prod-db"

  # 2) Interaktif dengan selection (tanpa --source, akan muncul pilihan)
  sfdbtools profile clone

  # 3) Non-interaktif (clone dengan nama baru dan host override)
  sfdbtools profile clone --quiet \
    --source "prod-db" \
    --name "prod-db-replica" \
    --host "10.0.0.6"

  # 4) Clone dengan override port dan enkripsi key berbeda
  sfdbtools profile clone \
    --source "prod-db" \
    --name "prod-db-standby" \
    --host "10.0.0.7" \
    --port 3307 \
    --profile-key "source-key-123" \
    --new-profile-key "target-key-456"

  # 5) Clone via environment variables
  SFDB_SOURCE_PROFILE_KEY='my-secret-key' \
  sfdbtools profile clone --quiet \
    --source "prod-db" \
    --name "prod-db-replica-2" \
    --host "10.0.0.8"`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := profile.ExecuteProfile(cmd, appdeps.Deps, "clone"); err != nil {
			appdeps.Deps.Logger.Error("profile clone gagal: " + err.Error())
		}
	},
}

func init() {
	flags.ProfileClone(CmdProfileClone)
}
