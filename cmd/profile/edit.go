package profilecmd

import (
	"sfDBTools/internal/flags"
	"sfDBTools/internal/profile"
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
)

var CmdProfileEdit = &cobra.Command{
	Use:   "edit",
	Short: "Mengubah konfigurasi profil database yang ada",
	Long: `Mengubah atau memperbarui nilai konfigurasi dalam file profil yang sudah ada.

Anda dapat mengubah satu atau beberapa parameter sekaligus (seperti host, password, user).
Command ini juga mendukung penggantian nama profil (rename) menggunakan flag --new-name.`,
	Example: `  # 1. Edit interaktif (memilih profil dan field yang akan diubah)
  sfdbtools profile edit

  # 2. Mengganti host dan port pada profil tertentu
  sfdbtools profile edit --profile "prod-db" --host "10.0.0.6" --port 3307

  # 3. Mengganti password user database
  sfdbtools profile edit --profile "prod-db" --password "new-password-2025"

  # 4. Mengganti nama profil (Rename)
  sfdbtools profile edit --profile "old-name" --new-name "new-name"

  # 5. Mengubah profil yang terenkripsi (memerlukan key lama jika ada)
  sfdbtools profile edit --profile "secure-db" --host "1.2.3.4" --profile-key "my-key"`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := profile.ExecuteProfile(cmd, types.Deps, "edit"); err != nil {
			types.Deps.Logger.Error("profile edit gagal: " + err.Error())
		}
	},
}

func init() {
	flags.ProfileEdit(CmdProfileEdit)
}
