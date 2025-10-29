// File : cmd/cmd_root.go
// Deskripsi : Root command untuk aplikasi sfDBTools
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package cmd

import (
	"fmt"
	"os"
	cmdcrypto "sfDBTools/cmd/cmd_crypto"
	cmddbscan "sfDBTools/cmd/cmd_db_scan"
	cmdprofile "sfDBTools/cmd/cmd_profile"
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
	// Import globals dan sub-command
)

// rootCmd merepresentasikan perintah dasar ketika tidak ada sub-command yang dipanggil
var rootCmd = &cobra.Command{
	Use:   "sfdbtools",
	Short: "SFDBTools: Database Backup and Management Utility",
	Long: `SFDBTools adalah utilitas manajemen dan backup MariaDB/MySQL.
Didesain untuk keandalan dan penggunaan di lingkungan produksi.`,

	// PersistentPreRunE akan dijalankan SEBELUM perintah apapun.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if types.Deps == nil || types.Deps.Config == nil || types.Deps.Logger == nil {
			return fmt.Errorf("dependensi belum di-inject. Pastikan untuk memanggil Execute(deps) dari main.go")
		}

		// Log bahwa perintah akan dieksekusi
		types.Deps.Logger.Infof("Memulai perintah: %s", cmd.Name())

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Silakan jalankan 'sfdbtools --help' untuk melihat perintah yang tersedia.")
	},
}

// Execute adalah fungsi eksekusi utama yang dipanggil dari main.go.
func Execute(deps *types.Dependencies) {
	// 1. INJEKSI DEPENDENSI
	types.Deps = deps

	// 2. Eksekusi perintah Cobra
	if err := rootCmd.Execute(); err != nil {
		if types.Deps != nil && types.Deps.Logger != nil {
			types.Deps.Logger.Fatalf("Gagal menjalankan perintah: %v", err)
		} else {
			fmt.Fprintf(os.Stderr, "Gagal menjalankan perintah: %v\n", err)
			os.Exit(1)
		}
	}
}

func init() {
	// Tambahkan sub-command yang sudah dibuat
	// Kita anggap 'versionCmd' ada di cmd/version.go
	rootCmd.AddCommand(versionCmd) // (Perlu diinisialisasi di cmd/version.go)
	rootCmd.AddCommand(cmdprofile.CmdProfileMain)
	rootCmd.AddCommand(cmddbscan.CmdDBScanMain)
	rootCmd.AddCommand(cmdcrypto.CmdCryptoMain)
}
