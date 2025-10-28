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

var CmdProfileShow = &cobra.Command{
	Use:   "show",
	Short: "Menampilkan detail profil pengguna",
	Long: `Perintah 'profile show' digunakan untuk menampilkan detail dari file profil pengguna yang ada.
Gunakan 'profile show --help' untuk informasi lebih lanjut tentang perintah ini.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			return fmt.Errorf("dependencies belum di-inject")
		}

		// Akses config dan logger dari dependency injection
		cfg := types.Deps.Config
		logger := types.Deps.Logger

		// Log dimulainya proses melihat profil
		logger.Info("Memulai melihat profil")

		// Membaca nilai dari flags
		ProfileShowOptions, err := parsing.ParsingShowProfile(cmd)
		if err != nil {
			logger.Errorf("Gagal memparsing flags: %v", err)
			return fmt.Errorf("gagal memparsing flags: %v", err)
		}
		// Inisialisasi service profile
		profileService := profile.NewService(cfg, logger, ProfileShowOptions)

		// Panggil method CreateProfile
		// Jalankan proses create
		if err := profileService.ShowProfile(); err != nil {
			if errors.Is(err, types.ErrUserCancelled) {
				return nil
			}
			return fmt.Errorf("create konfigurasi gagal: %w", err)
		}
		return nil
	},
}

func init() {
	flags.ProfileShow(CmdProfileShow)
}
