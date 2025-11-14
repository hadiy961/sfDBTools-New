package main

import (
	"fmt"
	"os"
	"sfDBTools/cmd"
	config "sfDBTools/internal/appconfig"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
	"time"

	applog "sfDBTools/internal/applog"

	"github.com/briandowns/spinner"
)

// Inisialisasi awal untuk Config dan Logger.
var cfg *config.Config
var appLogger applog.Logger

func main() {
	// Quiet mode for pipeline-friendly output
	quiet := false
	if v := os.Getenv(consts.ENV_QUIET); v != "" && v != "0" && v != "false" {
		quiet = true
	}

	// Deteksi jika yang dipanggil adalah perintah completion
	isCompletion := len(os.Args) > 1 && os.Args[1] == "completion"
	if isCompletion {
		quiet = true // pastikan tidak ada header/spinner yang tampil
	}

	if !quiet {
		ui.Headers("Main Menu")
	}

	// 1. Muat Konfigurasi (skip saat completion agar output bersih)
	var err error
	if !isCompletion {
		cfg, err = config.LoadConfigFromEnv()
		if err != nil {
			fmt.Fprintf(os.Stderr, "FATAL: Gagal memuat konfigurasi: %v\n", err)
			os.Exit(1)
		}
	}

	// 2. Inisialisasi Logger Kustom
	appLogger = applog.NewLogger()
	if quiet {
		// Route logs to stderr so stdout can be used for data piping
		appLogger.SetOutput(os.Stderr)
	}

	// 3. Inisialisasi koneksi database dari environment variables
	// In quiet mode atau saat completion, skip DB connection entirely
	if !quiet && !isCompletion {
		dbSpinner := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		dbSpinner.Suffix = " Menghubungkan ke database..."
		dbSpinner.Start()
		dbClient, err := database.ConnectToAppDatabase()
		dbSpinner.Stop()
		if err != nil {
			// Log error tapi jangan exit
			fmt.Println("✗ Gagal menghubungkan ke database")
			appLogger.Warn(fmt.Sprintf("Gagal menginisialisasi koneksi database: %v", err))
			return
		} else {
			// fmt.Println("✓ Berhasil terhubung ke database")
			defer dbClient.Close()
		}
	} else {
		// no-op
	}

	// 4. Buat objek dependensi untuk di-inject
	deps := &types.Dependencies{
		Config: cfg, // bisa nil saat completion, akan di-skip oleh PersistentPreRunE
		Logger: appLogger,
	}

	// 5. Jalankan perintah Cobra dengan dependensi
	cmd.Execute(deps)
}
