package defaultVal

import (
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/types"
)

// DefaultCleanupOptions mengembalikan opsi default untuk pembersihan backup.
func DefaultCleanupOptions() types.CleanupOptions {
	// Muat konfigurasi aplikasi untuk mendapatkan direktori konfigurasi
	cfg, _ := appconfig.LoadConfigFromEnv()

	opts := types.CleanupOptions{}

	opts.Enabled = cfg.Backup.Cleanup.Enabled
	opts.Days = cfg.Backup.Cleanup.Days
	opts.CleanupSchedule = cfg.Backup.Cleanup.Schedule // Setiap hari pukul 02:00
	opts.Pattern = ""

	// Fallback ke nilai default jika tidak ada konfigurasi
	return opts
}
