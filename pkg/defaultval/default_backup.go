package defaultVal

import (
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/types"
)

// DefaultBackupOptions mengembalikan default options untuk database backup
func DefaultBackupOptions(mode string) types.BackupDBOptions {
	// Muat konfigurasi aplikasi untuk mendapatkan direktori konfigurasi
	cfg, err := appconfig.LoadConfigFromEnv()

	opts := types.BackupDBOptions{}

	// Jika config tidak berhasil dimuat, kembalikan opsi default kosong (menghindari panic)
	if err != nil || cfg == nil {
		opts.Mode = mode
		return opts
	}

	// Compression Configuration
	opts.Compression.Enabled = cfg.Backup.Compression.Enabled
	opts.Compression.Type = cfg.Backup.Compression.Type
	opts.Compression.Level = cfg.Backup.Compression.Level
	// Encryption Configuration
	opts.Encryption.Enabled = cfg.Backup.Encryption.Enabled
	opts.Encryption.Key = cfg.Backup.Encryption.Key
	// Output Directory
	opts.OutputDir = cfg.Backup.Output.BaseDirectory
	// Capture GTID
	opts.CaptureGTID = cfg.Backup.Output.CaptureGtid
	// Dry Run
	opts.DryRun = false
	// Mode
	opts.Mode = mode

	// Exclude Filters (ambil langsung dari config agar tampil sebagai default di --help)
	opts.Filter.ExcludeSystem = cfg.Backup.Exclude.SystemDatabases
	opts.Filter.ExcludeUser = cfg.Backup.Exclude.User
	opts.Filter.ExcludeData = cfg.Backup.Exclude.Data
	opts.Filter.ExcludeEmpty = cfg.Backup.Exclude.Empty
	// Set keduanya: file dan daftar database (biarkan kosong jika memang tidak dikonfigurasi)
	opts.Filter.ExcludeDBFile = cfg.Backup.Exclude.File
	opts.Filter.ExcludeDatabases = cfg.Backup.Exclude.Databases

	// Include Filters
	opts.Filter.IncludeFile = cfg.Backup.Include.File
	opts.Filter.IncludeDatabases = cfg.Backup.Include.Databases
	opts.Force = false

	return opts
}
