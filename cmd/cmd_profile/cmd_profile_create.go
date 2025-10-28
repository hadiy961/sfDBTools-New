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
	Short: "Mengelola profil pengguna (generate, edit, delete, validate, show)",
	Long: `Perintah 'profile' digunakan untuk mengelola file profil pengguna.
Terdapat beberapa sub-perintah seperti create, edit, delete, validate, dan show.
Gunakan 'profile <sub-command> --help' untuk informasi lebih lanjut tentang masing-masing sub-perintah.`,
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
		profileService := profile.NewService(cfg, logger, ProfileCreateOptions)

		// Panggil method CreateProfile
		// Jalankan proses create
		if err := profileService.CreateProfile(); err != nil {
			// logger.Errorf("Proses pembuatan profil gagal: %v", err)
			return
		}
	},
}

func init() {
	flags.ProfileCreate(CmdProfileCreate)
}
