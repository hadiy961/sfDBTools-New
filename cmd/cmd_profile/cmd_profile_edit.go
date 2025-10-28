package cmdprofile

import (
	"errors"
	"fmt"
	"sfDBTools/internal/profile"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/flags"
	"sfDBTools/pkg/parsing"

	"github.com/spf13/cobra"
)

var CmdProfileEdit = &cobra.Command{
	Use:   "edit",
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

		// Log dimulainya proses edit profile
		logger.Info("Memulai proses pengeditan profil")

		// Membaca nilai dari flags
		ProfileEditOptions, err := parsing.ParsingEditProfile(cmd)
		if err != nil {
			logger.Errorf("Gagal memparsing flags: %v", err)
			return
		}
		fmt.Println("Interaktif  ", ProfileEditOptions.Interactive)
		// Inisialisasi service profile
		profileService := profile.NewService(cfg, logger, ProfileEditOptions)

		// Panggil method EditProfile
		// Jalankan proses edit
		if err := profileService.EditProfile(); err != nil {
			if errors.Is(err, types.ErrUserCancelled) {
				logger.Warn("Dibatalkan oleh pengguna.")
				return
			}
			// logger.Errorf("Edit konfigurasi gagal: %v", err)
		}
	},
}

func init() {
	flags.ProfileEdit(CmdProfileEdit)
}
