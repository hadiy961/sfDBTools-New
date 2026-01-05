package defaultVal

import (
	"os"
	dbscanmodel "sfDBTools/internal/app/dbscan/model"
	appconfig "sfDBTools/internal/services/config"
	"sfDBTools/pkg/consts"
)

// GetDefaultScanOptions mengembalikan default options untuk database scan
func GetDefaultScanOptions(mode string) dbscanmodel.ScanOptions {
	// Muat konfigurasi aplikasi untuk mendapatkan direktori konfigurasi
	cfg, _ := appconfig.LoadConfigFromEnv()

	opts := dbscanmodel.ScanOptions{}

	// Database Configuration
	opts.ProfileInfo.Path = os.Getenv(consts.ENV_SOURCE_PROFILE)

	// Encryption
	opts.Encryption.Key = os.Getenv(consts.ENV_SOURCE_PROFILE_KEY)

	// Database Selection
	if mode != "single" {
		opts.DatabaseList.UseFile = true
	}

	// Jika cfg nil (mis. environment SFDB_APPS_CONFIG tidak diset), gunakan fallback
	if cfg != nil {
		opts.DatabaseList.File = cfg.Backup.Include.File
	} else {
		// Fallback ke file list yang umum di-repo jika tidak ada konfigurasi
		opts.DatabaseList.File = "config/db_list.txt"
	}

	// Filter Options
	opts.ExcludeSystem = true

	// Output Options
	opts.DisplayResults = true
	opts.Background = false
	if opts.Background {
		opts.DisplayResults = false
	}
	opts.ShowOptions = true

	// Mode
	opts.Mode = mode

	return opts
}
