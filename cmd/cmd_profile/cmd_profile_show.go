package cmdprofile

import (
	"fmt"
	"sfDBTools/internal/profile"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/flags"
	"sfDBTools/pkg/parsing"

	"github.com/spf13/cobra"
)

var CmdProfileShow = &cobra.Command{
	Use:   "show",
	Short: "Tampilkan detail profil koneksi",
	Long: `Menampilkan informasi lengkap dari profil koneksi yang tersimpan, seperti host, port, user, database, dan status enkripsi password.
Gunakan flag --file/-f untuk menampilkan profil tertentu. Jika tidak memberikan --file, perintah dapat meminta pilihan secara interaktif (jika didukung).`,
	Example: `
	# Tampilkan detail profil bernama myprofile
	sfdbtools profile show --file myprofile

	# Tampilkan profil dan perlihatkan password secara jelas (jika diizinkan)
	sfdbtools profile show --file myprofile --reveal-password
`,
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
		profileService := profile.NewProfileService(cfg, logger, ProfileShowOptions)

		// Panggil method CreateProfile
		// Jalankan proses create
		if err := profileService.ShowProfile(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	flags.ProfileShow(CmdProfileShow)
}
