package defaultVal

import (
	"os"
	"sfDBTools/internal/services/config"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
)

// GetDefaultScanOptions mengembalikan default options untuk database scan
func GetDefaultScanOptions(mode string) types.ScanOptions {
	// Muat konfigurasi aplikasi untuk mendapatkan direktori konfigurasi
	cfg, _ := appconfig.LoadConfigFromEnv()

	opts := types.ScanOptions{}

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
