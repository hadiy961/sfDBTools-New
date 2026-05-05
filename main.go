package main

import (
	"context"
	"fmt"
	"os"
	"sfdbtools/cmd"
	"sfdbtools/internal/autoupdate"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/services/config"
	"sfdbtools/internal/services/notify"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/progress"
	"time"

	applog "sfdbtools/internal/services/log"
	"github.com/fatih/color"
)

// Inisialisasi awal untuk Config dan Logger.
var cfg *appconfig.Config
var appLogger applog.Logger

func main() {
	// Bootstrap runtime mode dari parameter (tanpa env).
	runtimecfg.BootstrapFromArgs(os.Args[1:])
	quiet := runtimecfg.IsQuiet()

	// Deteksi jika yang dipanggil adalah perintah completion
	isCompletion := len(os.Args) > 1 && os.Args[1] == "completion"
	isVersion := len(os.Args) > 1 && os.Args[1] == "version"
	isUpdate := len(os.Args) > 1 && os.Args[1] == "update"
	isInit := len(os.Args) > 1 && os.Args[1] == "init"
	if isCompletion {
		quiet = true // pastikan tidak ada header/spinner yang tampil
		runtimecfg.SetQuiet(true)
	}
	if isVersion {
		quiet = true // output versi sebaiknya bersih untuk scripting
		runtimecfg.SetQuiet(true)
	}
	if isInit {
		// Init harus interaktif, jangan set quiet
		runtimecfg.SetQuiet(false)
	}
	if isUpdate {
		quiet = true // update sebaiknya bersih dan tidak perlu header
		runtimecfg.SetQuiet(true)
	}

	// Auto-update dijalankan sebelum load config agar tidak tergantung config.yaml.
	// Skip untuk completion/version/update.
	if !isCompletion && !isVersion && !isUpdate {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		// Logger aman walau cfg nil.
		tmpLogger := applog.NewLogger(nil)

		// Tampilkan spinner hanya jika auto-update aktif dan tidak dalam quiet.
		var sp *progress.Spinner
		if autoupdate.AutoUpdateEnabled() && !runtimecfg.IsQuiet() {
			sp = progress.NewSpinnerWithElapsed("Cek update")
			sp.Start()
		}
		if err := autoupdate.MaybeAutoUpdate(ctx, tmpLogger); err != nil {
			// Jangan silent-fail: biasanya error permission (/usr/bin) atau koneksi.
			tmpLogger.Warnf("Auto-update gagal: %v", err)
		}
		if sp != nil {
			sp.Stop()
		}
	}

	if !quiet {
		print.PrintAppHeader("Main Menu")
	}

	// 1. Muat Konfigurasi (skip saat completion/version agar output bersih)
	var err error
	if !isCompletion && !isVersion && !isUpdate {
		cfg, err = appconfig.LoadConfigFromEnv()
		if err != nil {
			if isInit {
				// Jika init, jangan fatal error, biarkan init yang menghandle/memperbaiki.
				fmt.Printf("%s Konfigurasi tidak valid atau hilang, masuk ke mode inisialisasi...\n", color.YellowString("[WARN]"))
			} else {
				fmt.Fprintf(os.Stderr, "FATAL: Gagal memuat konfigurasi: %v\n", err)
				fmt.Println("Gunakan 'sfdbtools init' untuk memperbaiki konfigurasi.")
				os.Exit(1)
			}
		}
		if cfg != nil {
			// Konfigurasi path key file MariaDB untuk derive master key env terenkripsi.
			// Jika kosong, akan fallback ke default internal.
			crypto.SetMariaDBKeyFilePath(cfg.Mariadb.KeyMariaNBCFile)
		}
	}

	// 2. Inisialisasi Logger Kustom
	appLogger = applog.NewLogger(cfg)

	// 3. Inisialisasi koneksi database dari environment variables
	// In quiet mode atau saat completion, skip DB connection entirely
	// if !quiet && !isCompletion {
	// 	dbSpinner := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	// 	dbSpinner.Suffix = " Menghubungkan ke database..."
	// 	dbSpinner.Start()
	// 	dbClient, err := database.ConnectToAppDatabase()
	// 	dbSpinner.Stop()
	// 	if err != nil {
	// 		// Log error tapi jangan exit
	// 		fmt.Println("✗ Gagal menghubungkan ke database")
	// 		appLogger.Warn(fmt.Sprintf("Gagal menginisialisasi koneksi database: %v", err))
	// 		return
	// 	} else {
	// 		// fmt.Println("✓ Berhasil terhubung ke database")
	// 		defer dbClient.Close()
	// 	}
	// } else {
	// 	// no-op
	// }

	// 4. Buat objek dependensi untuk di-inject
	var notifyService *notify.Service
	if cfg != nil {
		notifyService = notify.NewService(&cfg.Notify, appLogger)
	} else {
		notifyService = notify.NewService(nil, appLogger)
	}

	deps := &appdeps.Dependencies{
		Config:        cfg, // bisa nil saat completion, akan di-skip oleh PersistentPreRunE
		Logger:        appLogger,
		NotifyService: notifyService,
	}

	// 5. Jalankan perintah Cobra dengan dependensi
	cmd.Execute(deps)
}
