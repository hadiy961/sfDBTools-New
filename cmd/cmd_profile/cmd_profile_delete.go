package cmdprofile

import (
	"errors"
	"fmt"

	"sfDBTools/internal/profile"
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
)

var CmdProfileDelete = &cobra.Command{
	Use:   "delete",
	Short: "Menghapus file konfigurasi database yang sudah ada",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			return fmt.Errorf("dependencies belum di-inject")
		}
		// Akses config dan logger dari dependency injection
		cfg := types.Deps.Config
		logger := types.Deps.Logger

		// Log dimulainya proses delete config
		logger.Info("Memulai proses menghapus konfigurasi database...")

		// Buat service dbconfig tanpa perlu state khusus
		service := profile.NewService(cfg, logger, nil)

		// Jalankan proses delete dengan prompt konfirmasi
		if err := service.PromptDeleteProfile(); err != nil {
			if errors.Is(err, types.ErrUserCancelled) {
				logger.Warn("Dibatalkan oleh pengguna.")
				return nil
			}
			return fmt.Errorf("penghapusan profil gagal: %w", err)
		}
		return nil
	},
}
