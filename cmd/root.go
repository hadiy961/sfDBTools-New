// File : cmd/cmd_root.go
// Deskripsi : Root command untuk aplikasi sfdbtools
// Author : Hadiyatna Muflihun
// Tanggal : 3 Oktober 2024
// Last Modified : 26 Januari 2026
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
	backupcmd "sfdbtools/cmd/backup"
	cleanupcmd "sfdbtools/cmd/cleanup"
	copycmd "sfdbtools/cmd/copy"
	cryptocmd "sfdbtools/cmd/crypto"
	dbusercmd "sfdbtools/cmd/dbuser"
	profilecmd "sfdbtools/cmd/profile"
	restorecmd "sfdbtools/cmd/restore"
	scriptcmd "sfdbtools/cmd/script"
	notifycmd "sfdbtools/cmd/notify"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/shared/sanitize"
	"sfdbtools/internal/ui/menu"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/progress"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/app/settings"
	"sfdbtools/internal/autoupdate"
	"sfdbtools/internal/app/sync"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// rootCmd merepresentasikan perintah dasar ketika tidak ada sub-command yang dipanggil
var rootCmd = &cobra.Command{
	Use:   "sfdbtools",
	Short: "SFDBTools: Database Backup and Management Utility",
	Long: `SFDBTools adalah utilitas manajemen dan backup MariaDB/MySQL.
Didesain untuk keandalan dan penggunaan di lingkungan produksi.`,

	// PersistentPreRunE akan dijalankan SEBELUM perintah apapun.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Apply global runtime flags seawal mungkin untuk memastikan mode quiet konsisten.
		quietFlag, _ := cmd.Flags().GetBool("quiet")
		quiteFlag, _ := cmd.Flags().GetBool("quite")
		if quietFlag || quiteFlag {
			runtimecfg.SetQuiet(true)
		}

		// Skip dependensi dan logging untuk perintah completion agar output bersih
		if cmd.Name() == "completion" || cmd.HasParent() && cmd.Parent().Name() == "completion" {
			return nil
		}
		// Version, update, dan init harus bisa jalan tanpa config (mis. sebelum instalasi config.yaml)
		if cmd.Name() == "version" || cmd.Name() == "update" || cmd.Name() == "init" {
			return nil
		}

		if appdeps.Deps == nil || appdeps.Deps.Config == nil || appdeps.Deps.Logger == nil {
			return fmt.Errorf("dependensi belum di-inject. Pastikan untuk memanggil Execute(deps) dari main.go")
		}

		// [VERIFIKASI INIT] Cek apakah database sudah diinisialisasi
		if !database.IsInitialized(appdeps.Deps.Config.Storage.LocalDB) {
			fmt.Println(color.RedString("\n[ERROR] Aplikasi belum diinisialisasi!"))
			fmt.Println("Silakan jalankan perintah berikut untuk setup awal:")
			fmt.Println(color.CyanString("  sfdbtools init\n"))
			os.Exit(1)
		}

		// [INIT SQLITE] Inisialisasi instance database global untuk digunakan di modul lain
		if err := database.InitSQLite(appdeps.Deps.Config.Storage.LocalDB); err != nil {
			return fmt.Errorf("gagal menginisialisasi database sqlite: %v", err)
		}

		// [SQLITE SYNC] Override konfigurasi YAML dengan nilai dari database
		settings.SyncConfig(appdeps.Deps.Config)

		// [STARTUP NETWORK SEQUENCE]
		checkInternet := database.GetSettingBool("check_internet_startup")
		autoUpdate := database.GetSettingBool("auto_update_enabled")
		autoSync := database.GetSettingBool("sync_auto") && database.GetSettingBool("sync_enabled")

		// Hanya jalankan blok network startup jika internet check aktif
		if checkInternet && (autoUpdate || autoSync) {
			// 1. Auto Update Check
			if autoUpdate {
				sp := progress.NewSpinner("Cek update")
				sp.Start()
				_ = autoupdate.MaybeAutoUpdate(cmd.Context(), appdeps.Deps.Logger)
				sp.Stop()
			}

			// 2. Unified Auto Sync
			if autoSync {
				ctx, cancel := context.WithTimeout(cmd.Context(), 45*time.Second)
				_ = sync.PerformFullSync(ctx, appdeps.Deps.Config.General.ClientCode, false)
				cancel()
			}
		}

		// Log bahwa perintah akan dieksekusi, termasuk argumen (tanpa membocorkan secret).
		bin := filepath.Base(os.Args[0])
		argsSafe := sanitize.Args(os.Args[1:])
		argLine := strings.Join(argsSafe, " ")
		if strings.TrimSpace(argLine) == "" {
			argLine = "-"
		}
		appdeps.Deps.Logger.Infof("Memulai perintah: %s | argv: %s %s", cmd.CommandPath(), bin, argLine)

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		// Jika ada argumen/flags, skip menu interaktif
		if len(args) > 0 || cmd.Flags().NFlag() > 0 {
			fmt.Println("Silakan jalankan 'sfdbtools --help' untuk melihat perintah yang tersedia.")
			print.PrintSeparator()
			cmd.Help()
			return
		}

		// Tampilkan menu interaktif
		if err := menu.ShowInteractiveMenu(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
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
	// Global flags (parameter-only): tanpa perlu SFDB_QUIET.
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Output lebih senyap (tanpa spinner/header), cocok untuk pipeline")
	rootCmd.PersistentFlags().Bool("quite", false, "Alias (deprecated) untuk --quiet")
	_ = rootCmd.PersistentFlags().MarkHidden("quite")
	_ = rootCmd.PersistentFlags().MarkDeprecated("quite", "gunakan --quiet")

	// Tambahkan sub-command yang sudah dibuat
	// Kita anggap 'versionCmd' ada di cmd/version.go
	rootCmd.AddCommand(versionCmd) // (Perlu diinisialisasi di cmd/version.go)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(profilecmd.CmdProfileMain)
	rootCmd.AddCommand(cryptocmd.CmdCryptoMain)
	rootCmd.AddCommand(scriptcmd.CmdScriptMain)
	rootCmd.AddCommand(cleanupcmd.CmdCleanupMain)
	rootCmd.AddCommand(backupcmd.CmdBackupMain)
	rootCmd.AddCommand(restorecmd.CmdRestore)
	rootCmd.AddCommand(copycmd.CmdCopyMain)
	rootCmd.AddCommand(dbusercmd.CmdDBUserMain)
	rootCmd.AddCommand(notifycmd.CmdNotifyMain)
	rootCmd.AddCommand(completionCmd)
}
