package defaultVal

import (
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/helper"
)

// DefaultBackupOptions mengembalikan default options untuk database backup
func DefaultBackupOptions(mode string) types_backup.BackupDBOptions {
	// Muat konfigurasi aplikasi untuk mendapatkan direktori konfigurasi
	cfg, err := appconfig.LoadConfigFromEnv()

	opts := types_backup.BackupDBOptions{}

	// Jika config tidak berhasil dimuat, kembalikan opsi default kosong (menghindari panic)
	if err != nil || cfg == nil {
		opts.Mode = mode
		return opts
	}

	// Compression Configuration - derive Enabled from Type (false if type is "none" or empty)
	opts.Compression.Type = cfg.Backup.Compression.Type
	opts.Compression.Level = cfg.Backup.Compression.Level
	opts.Compression.Enabled = cfg.Backup.Compression.Type != "" && cfg.Backup.Compression.Type != "none"
	// Encryption Configuration - derive Enabled from Key (true if key is not empty)
	opts.Encryption.Key = cfg.Backup.Encryption.Key
	opts.Encryption.Enabled = cfg.Backup.Encryption.Key != ""
	// Output Directory Configuration
	// Note: OutputDir ditampilkan dengan structure pattern yang sudah di-substitute dengan timestamp saat ini
	// Contoh: /media/ArchiveDB/{year}{month}{day}/ menjadi /media/ArchiveDB/20251205/
	// Hostname TIDAK di-include di directory (hanya untuk filename)
	outputDir, _ := helper.GenerateBackupDirectory(
		cfg.Backup.Output.BaseDirectory,
		cfg.Backup.Output.Structure.Pattern,
		"", // hostname tidak diperlukan untuk preview directory
	)
	opts.OutputDir = outputDir
	// Capture GTID (hanya untuk combined mode)
	if mode == "combined" {
		opts.CaptureGTID = cfg.Backup.Replication.CaptureGtid
	} else {
		opts.CaptureGTID = false
	}
	// Exclude User - ambil dari config backup.exclude.user
	opts.ExcludeUser = cfg.Backup.Exclude.User
	// Dry Run
	opts.DryRun = false
	// Mode
	opts.Mode = mode

	// Exclude Filters (ambil langsung dari config agar tampil sebagai default di --help)
	opts.Filter.ExcludeSystem = cfg.Backup.Exclude.SystemDatabases
	opts.Filter.ExcludeData = cfg.Backup.Exclude.Data
	opts.Filter.ExcludeEmpty = cfg.Backup.Exclude.Empty
	// Set keduanya: file dan daftar database (biarkan kosong jika memang tidak dikonfigurasi)
	opts.Filter.ExcludeDBFile = cfg.Backup.Exclude.File
	opts.Filter.ExcludeDatabases = cfg.Backup.Exclude.Databases

	// Include Filters
	opts.Filter.IncludeFile = cfg.Backup.Include.File
	opts.Filter.IncludeDatabases = cfg.Backup.Include.Databases
	// Include linked/companion databases (khusus single backup)
	opts.IncludeDmart = cfg.Backup.Include.IncludeDmart
	opts.IncludeTemp = cfg.Backup.Include.IncludeTemp
	opts.IncludeArchive = cfg.Backup.Include.IncludeArchive
	opts.Force = false

	return opts
}
