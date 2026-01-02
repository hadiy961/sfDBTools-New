// File : cmd/cmd_root.go
// Deskripsi : Root command untuk aplikasi sfDBTools
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2026-01-02
package cmd

import (
	"fmt"
	"os"
	backupcmd "sfDBTools/cmd/backup"
	cleanupcmd "sfDBTools/cmd/cleanup"
	cryptocmd "sfDBTools/cmd/crypto"
	dbscancmd "sfDBTools/cmd/dbscan"
	profilecmd "sfDBTools/cmd/profile"
	restorecmd "sfDBTools/cmd/restore"
	scriptcmd "sfDBTools/cmd/script"
	appdeps "sfDBTools/internal/deps"
	"sfDBTools/pkg/ui"

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
		// Skip dependensi dan logging untuk perintah completion agar output bersih
		if cmd.Name() == "completion" || cmd.HasParent() && cmd.Parent().Name() == "completion" {
			return nil
		}
		// Version harus bisa jalan tanpa config (mis. sebelum instalasi config.yaml)
		if cmd.Name() == "version" {
			return nil
		}
		if appdeps.Deps == nil || appdeps.Deps.Config == nil || appdeps.Deps.Logger == nil {
			return fmt.Errorf("dependensi belum di-inject. Pastikan untuk memanggil Execute(deps) dari main.go")
		}

		// Log bahwa perintah akan dieksekusi
		appdeps.Deps.Logger.Infof("Memulai perintah: %s", cmd.Name())

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Silakan jalankan 'sfdbtools --help' untuk melihat perintah yang tersedia.")
		ui.PrintSeparator()
		cmd.Help()
	},
}

// Execute adalah fungsi eksekusi utama yang dipanggil dari main.go.
func Execute(deps *appdeps.Dependencies) {
	// 1. INJEKSI DEPENDENSI
	appdeps.Deps = deps

	// 2. Eksekusi perintah Cobra
	if err := rootCmd.Execute(); err != nil {
		if appdeps.Deps != nil && appdeps.Deps.Logger != nil {
			appdeps.Deps.Logger.Fatalf("Gagal menjalankan perintah: %v", err)
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
	rootCmd.AddCommand(profilecmd.CmdProfileMain)
	rootCmd.AddCommand(dbscancmd.CmdDBScanMain)
	rootCmd.AddCommand(cryptocmd.CmdCryptoMain)
	rootCmd.AddCommand(scriptcmd.CmdScriptMain)
	rootCmd.AddCommand(cleanupcmd.CmdCleanupMain)
	rootCmd.AddCommand(backupcmd.CmdBackupMain)
	rootCmd.AddCommand(restorecmd.CmdRestore)
	rootCmd.AddCommand(completionCmd)
}
