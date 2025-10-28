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
	Short: "Edit profil koneksi database yang sudah ada",
	Long: `Mengubah nilai pada profil koneksi yang sudah tersimpan.
Anda dapat menjalankan secara non-interaktif dengan flag yang menunjuk field yang ingin diubah, atau menjalankan mode interaktif untuk mengedit nilai langkah demi langkah.
Pastikan profil ditentukan melalui --file/-f agar profil yang tepat dapat ditemukan; untuk mengganti nama gunakan --new-name/-N.`,
	Example: `
	# Edit host pada profil dengan file/nama myprofile
	sfdbtools profile edit --file myprofile --host 192.168.1.10

	# Edit interaktif (akan menanyakan field yang ingin diubah)
	sfdbtools profile edit --file myprofile --interactive
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

		// Log dimulainya proses edit profile
		logger.Info("Memulai proses pengeditan profil")

		// Membaca nilai dari flags
		ProfileEditOptions, err := parsing.ParsingEditProfile(cmd)
		if err != nil {
			logger.Errorf("Gagal memparsing flags: %v", err)
			return
		}
		// Debug: tunjukkan bahwa perintah terpanggil dan nilai flag
		fmt.Println("DEBUG: profile edit handler invoked")
		fmt.Printf("DEBUG: parsed options: interactive=%v, file=%s, new-name=%s\n",
			ProfileEditOptions.Interactive,
			ProfileEditOptions.ProfileInfo.Path,
			ProfileEditOptions.NewName,
		)
		// Tampilkan juga ke logger jika tersedia (logger mungkin tidak ke-stdout bergantung pada config)
		logger.Infof("ProfileEdit options: interactive=%v, file=%s, new-name=%s",
			ProfileEditOptions.Interactive,
			ProfileEditOptions.ProfileInfo.Path,
			ProfileEditOptions.NewName,
		)
		// Inisialisasi service profile
		profileService := profile.NewService(cfg, logger, ProfileEditOptions)

		// Panggil method EditProfile
		// Jalankan proses edit
		if err := profileService.EditProfile(); err != nil {
			if errors.Is(err, types.ErrUserCancelled) {
				logger.Warn("Dibatalkan oleh pengguna.")
				return
			}
			// Laporkan error lain agar terlihat di konsol/log
			logger.Errorf("Edit konfigurasi gagal: %v", err)
			return
		}
	},
}

func init() {
	flags.ProfileEdit(CmdProfileEdit)
}
