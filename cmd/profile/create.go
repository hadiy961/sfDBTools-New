package profilecmd

import (
	appdeps "sfDBTools/internal/deps"
	"sfDBTools/internal/flags"
	"sfDBTools/internal/profile"

	"github.com/spf13/cobra"
)

var CmdProfileCreate = &cobra.Command{
	Use:   "create",
	Short: "Membuat profil koneksi database baru",
	Long: `Membuat profil koneksi database baru untuk digunakan oleh sfDBTools.

Command ini akan menyimpan kredensial dan konfigurasi koneksi ke dalam file profil.
Profil ini nantinya digunakan untuk operasi backup, restore, dan dbscan tanpa perlu memasukkan ulang kredensial.

Fitur:
  - Mode Wizard: Jalankan tanpa flag untuk panduan interaktif langkah demi langkah.
  - Enkripsi: Opsi untuk mengenkripsi password atau seluruh profil menggunakan key.
  - Validasi: Melakukan tes koneksi otomatis untuk memastikan profil valid sebelum disimpan.
  - Custom Path: Simpan profil di lokasi khusus dengan --output-dir.`,
	Example: `  # 1. Mode Interaktif (Wizard)
  sfdbtools profile create

  # 2. Mode One-Liner (Langsung dengan flag)
  sfdbtools profile create \
    --profile "prod-db-primary" \
    --host "10.0.0.5" \
    --port 3306 \
    --user "admin" \
    --password "s3cr3t" \
    --database "main_app" \
    --description "Database utama produksi"

  # 3. Membuat profil dengan enkripsi custom
  sfdbtools profile create \
    --profile "secure-db" \
    --host "localhost" \
    --user "root" \
    --password "toor" \
    --profile-key "my-secret-key-123"

  # 4. Menyimpan di direktori khusus
  sfdbtools profile create \
    --profile "local-dev" \
    --output-dir "./configs"`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := profile.ExecuteProfile(cmd, appdeps.Deps, "create"); err != nil {
			appdeps.Deps.Logger.Error("profile create gagal: " + err.Error())
		}
	},
}

func init() {
	flags.ProfileCreate(CmdProfileCreate)
}
