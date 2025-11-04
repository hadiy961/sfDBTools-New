package defaultVal

import (
	"os"
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"strconv"
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

	// Target Database (untuk menyimpan hasil scan)
	opts.TargetDB.Host = os.Getenv(consts.ENV_DB_HOST)
	if opts.TargetDB.Host == "" {
		opts.TargetDB.Host = "localhost"
	}

	portStr := os.Getenv(consts.ENV_DB_PORT)
	if portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			opts.TargetDB.Port = port
		} else {
			opts.TargetDB.Port = 3306
		}
	} else {
		opts.TargetDB.Port = 3306
	}

	opts.TargetDB.User = os.Getenv(consts.ENV_DB_USER)
	if opts.TargetDB.User == "" {
		opts.TargetDB.User = "root"
	}

	opts.TargetDB.Password = os.Getenv(consts.ENV_DB_PASSWORD)
	opts.TargetDB.Database = os.Getenv(consts.ENV_DB_NAME)
	if opts.TargetDB.Database == "" {
		opts.TargetDB.Database = "sfdbtools"
	}

	// Output Options
	opts.DisplayResults = true
	opts.SaveToDB = true
	opts.Background = false
	if opts.Background {
		opts.DisplayResults = false
	}
	opts.ShowOptions = true

	// Mode
	opts.Mode = mode

	return opts
}
