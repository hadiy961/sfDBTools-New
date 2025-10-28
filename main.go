package main

import (
	"fmt"
	"os"
	"sfDBTools/cmd"
	config "sfDBTools/internal/appconfig"
	"sfDBTools/internal/types"
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
	ui.Headers("Main Menu")

	// 1. Muat Konfigurasi
	var err error
	cfg, err = config.LoadConfigFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: Gagal memuat konfigurasi: %v\n", err)
		os.Exit(1)
	}

	// 2. Inisialisasi Logger Kustom
	appLogger = applog.NewLogger()

	// 3. Inisialisasi koneksi database dari environment variables
	dbSpinner := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	dbSpinner.Suffix = " Menghubungkan ke database..."
	dbSpinner.Start()

	dbClient, err := database.ConnectToAppDatabase()
	dbSpinner.Stop()

	if err != nil {
		// Log error tapi jangan exit, karena tidak semua command membutuhkan database
		fmt.Println("✗ Gagal menghubungkan ke database")
		appLogger.Warn(fmt.Sprintf("Gagal menginisialisasi koneksi database: %v", err))
		return
	} else {
		fmt.Println("✓ Berhasil terhubung ke database")
		defer dbClient.Close()
	}

	// 4. Buat objek dependensi untuk di-inject
	deps := &types.Dependencies{
		Config: cfg,
		Logger: appLogger,
	}

	// 5. Jalankan perintah Cobra dengan dependensi
	cmd.Execute(deps)
}
