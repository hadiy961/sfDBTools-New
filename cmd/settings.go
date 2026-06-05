package cmd

import (
	"sfdbtools/internal/app/settings"
	"sfdbtools/internal/cli/deps"

	"github.com/spf13/cobra"
)

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Kelola konfigurasi aplikasi yang tersimpan di database",
	Long:  `Modul untuk melihat dan mengubah pengaturan aplikasi seperti notifikasi, backup, dan sinkronisasi cloud.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Jika tidak ada subcommand, tampilkan menu utama
		s := settings.NewService(deps.Deps)
		return s.RunMenu()
	},
}

var settingsViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Lihat semua konfigurasi aplikasi",
	Run: func(cmd *cobra.Command, args []string) {
		s := settings.NewService(deps.Deps)
		s.ViewAllSettings()
	},
}

var settingsEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Ubah konfigurasi aplikasi secara interaktif",
	Run: func(cmd *cobra.Command, args []string) {
		s := settings.NewService(deps.Deps)
		s.EditSettingsMenu()
	},
}

var settingsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Kembalikan semua pengaturan ke nilai default",
	Run: func(cmd *cobra.Command, args []string) {
		s := settings.NewService(deps.Deps)
		s.ResetSettings()
	},
}

func init() {
	settingsCmd.AddCommand(settingsViewCmd)
	settingsCmd.AddCommand(settingsEditCmd)
	settingsCmd.AddCommand(settingsResetCmd)
	rootCmd.AddCommand(settingsCmd)
}
