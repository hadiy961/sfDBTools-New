package cmdprofile

import (
	"fmt"
	"sfDBTools/internal/profile"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/flags"
	"sfDBTools/pkg/parsing"

	"github.com/spf13/cobra"
)

var CmdProfileCreate = &cobra.Command{
	Use:   "create",
	Short: "Buat profil koneksi database baru",
	Long: `Membuat profil koneksi database baru dan menyimpannya ke konfigurasi lokal.
Anda dapat memberikan detail melalui flag (mis. --profil/-n untuk nama profil, --host/-H, --port/-P, --user/-U).
Tersedia juga opsi seperti --output-dir/-o untuk lokasi penyimpanan, --key/-k untuk kunci enkripsi, dan --interactive/-i untuk mode interaktif.
Jika flag tidak lengkap, proses dapat meminta input secara interaktif.`,
	Example: `
	# Buat profil non-interaktif (gunakan flag --profil/-n untuk nama profil)
	sfdbtools profile create --profil myprofile --host 127.0.0.1 --port 3306 --user root

	# Buat profil interaktif (akan menanyakan nilai yang belum diisi)
	sfdbtools profile create --interactive
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		// Akses config dan logger dari dependency injection
		cfg := types.Deps.Config
		logger := types.Deps.Logger

		// Log dimulainya proses create profile
		logger.Info("Memulai proses pembuatan profil")

		// Membaca nilai dari flags
		ProfileCreateOptions, err := parsing.ParsingCreateProfile(cmd, logger)
		if err != nil {
			logger.Errorf("Gagal memparsing flags: %v", err)
			return
		}

		// Inisialisasi service profile
		profileService := profile.NewProfileService(cfg, logger, ProfileCreateOptions)

		// Panggil method CreateProfile
		// Jalankan proses create
		if err := profileService.CreateProfile(); err != nil {
			return
		}
	},
}

func init() {
	flags.ProfileCreate(CmdProfileCreate)
}
