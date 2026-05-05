// File : cmd/init.go
// Deskripsi : Command untuk inisialisasi aplikasi dan migrasi database
// Author : Antigravity
// Tanggal : 30 April 2026

package cmd

import (
	"fmt"
	"os"
	"sfdbtools/internal/app/initialize"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/shared/database"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var cmdInit = &cobra.Command{
	Use:   "init",
	Short: "Inisialisasi aplikasi dan migrasi konfigurasi ke database",
	Long: `Perintah ini digunakan untuk menyiapkan database SQLite lokal, 
melakukan migrasi setting dari config.yaml, dan menghubungkan ke Supabase.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Perintah init boleh jalan meskipun config belum lengkap,
		// tapi kita butuh setidaknya logger.
		if appdeps.Deps == nil || appdeps.Deps.Logger == nil {
			fmt.Println("Error: Logger belum siap.")
			os.Exit(1)
		}

		// [PREVENTIVE CHECK]
		// Jika database sudah terinisialisasi dan config.yaml valid (punya client_code),
		// maka blokir jalannya init.
		isInitialized := false
		if appdeps.Deps.Config != nil && appdeps.Deps.Config.Storage.LocalDB != "" {
			isInitialized = database.IsInitialized(appdeps.Deps.Config.Storage.LocalDB)
		} else {
			// Jika config nil, tapi kita bisa tebak path default SQLite
			defaultDBPath := "/etc/sfDBTools/sfdbtools.db"
			isInitialized = database.IsInitialized(defaultDBPath)
		}

		if isInitialized && appdeps.Deps.Config != nil && appdeps.Deps.Config.General.ClientCode != "" {
			fmt.Println(color.YellowString("\n[INFO] Aplikasi sudah diinisialisasi sebelumnya."))
			fmt.Println("Pengaturan operasional sekarang dikelola melalui database SQLite.")
			fmt.Println("\nUntuk merubah konfigurasi, silakan gunakan perintah:")
			fmt.Println(color.CyanString("  sfdbtools settings\n"))
			return
		}

		// Jalankan service inisialisasi
		svc := initialize.NewService(appdeps.Deps)
		if err := svc.RunWizard(); err != nil {
			appdeps.Deps.Logger.Errorf("Gagal menjalankan inisialisasi: %v", err)
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n[SUCCESS] Inisialisasi selesai!")
	},
}

func init() {
	rootCmd.AddCommand(cmdInit)
}
