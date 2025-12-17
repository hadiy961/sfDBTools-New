package defaultVal

import (
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/types"
)

// DefaultCleanupOptions mengembalikan opsi default untuk pembersihan backup.
// Aman dipanggil saat inisialisasi (init) karena tidak akan panic jika konfigurasi belum tersedia.
func DefaultCleanupOptions() types.CleanupOptions {
	// Inisialisasi dengan nilai paling aman (disabled) agar tidak ada operasi berbahaya terjadi tanpa konfigurasi.
	opts := types.CleanupOptions{
		Enabled:         false,
		Days:            0,
		CleanupSchedule: "",
		Pattern:         "",
	}

	// Muat konfigurasi aplikasi (abaikan error; jika gagal kita pakai nilai aman di atas).
	cfg, err := appconfig.LoadConfigFromEnv()
	if err != nil || cfg == nil {
		return opts
	}

	// Set nilai dari konfigurasi jika tersedia.
	opts.Enabled = cfg.Backup.Cleanup.Enabled
	opts.Days = cfg.Backup.Cleanup.Days
	opts.CleanupSchedule = cfg.Backup.Cleanup.Schedule
	// Pattern dibiarkan kosong (bisa diubah via flag saat runtime)

	return opts
}
