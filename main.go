package main

import (
	"context"
	"fmt"
	"os"
	"sfDBTools/cmd"
	"sfDBTools/internal/autoupdate"
	appdeps "sfDBTools/internal/cli/deps"
	config "sfDBTools/internal/services/config"
	"sfDBTools/pkg/runtimecfg"
	"sfDBTools/pkg/ui"
	"time"

	applog "sfDBTools/internal/services/log"
)

// Inisialisasi awal untuk Config dan Logger.
var cfg *config.Config
var appLogger applog.Logger

func main() {
	// Bootstrap runtime mode dari parameter (tanpa env).
	runtimecfg.BootstrapFromArgs(os.Args[1:])
	quiet := runtimecfg.IsQuiet() || runtimecfg.IsDaemon()

	// Deteksi jika yang dipanggil adalah perintah completion
	isCompletion := len(os.Args) > 1 && os.Args[1] == "completion"
	isVersion := len(os.Args) > 1 && os.Args[1] == "version"
	isUpdate := len(os.Args) > 1 && os.Args[1] == "update"
	if isCompletion {
		quiet = true // pastikan tidak ada header/spinner yang tampil
		runtimecfg.SetQuiet(true)
	}
	if isVersion {
		quiet = true // output versi sebaiknya bersih untuk scripting
		runtimecfg.SetQuiet(true)
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

		// Tampilkan spinner hanya jika auto-update aktif dan tidak dalam quiet/daemon.
		var sp *ui.SpinnerWithElapsed
		if autoupdate.AutoUpdateEnabled() && !(runtimecfg.IsQuiet() || runtimecfg.IsDaemon()) {
			sp = ui.NewSpinnerWithElapsed("Cek update")
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
		ui.Headers("Main Menu")
	}

	// 1. Muat Konfigurasi (skip saat completion/version agar output bersih)
	var err error
	if !isCompletion && !isVersion && !isUpdate {
		cfg, err = config.LoadConfigFromEnv()
		if err != nil {
			fmt.Fprintf(os.Stderr, "FATAL: Gagal memuat konfigurasi: %v\n", err)
			os.Exit(1)
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
	deps := &appdeps.Dependencies{
		Config: cfg, // bisa nil saat completion, akan di-skip oleh PersistentPreRunE
		Logger: appLogger,
	}

	// 5. Jalankan perintah Cobra dengan dependensi
	cmd.Execute(deps)
}
