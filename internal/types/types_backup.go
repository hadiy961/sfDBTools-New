package types

import (
	"time"
)

// BackupDBOptions menyimpan opsi konfigurasi untuk proses backup database.
type BackupDBOptions struct {
	Filter      FilterOptions
	Profile     ProfileInfo
	Compression CompressionOptions
	Encryption  EncryptionOptions
	Cleanup     CleanupOptions
	DryRun      bool
	OutputDir   string
	Mode        string         // "separate" atau "combined"
	Force       bool           // Tampilkan opsi backup sebelum eksekusi
	File        BackupFileInfo // Nama file backup lengkap dengan ekstensi
	Entry       BackupEntryConfig
	CaptureGTID bool // Tangkap informasi GTID saat backup
}

// BackupConfig untuk konfigurasi backup entry point
type BackupEntryConfig struct {
	HeaderTitle string
	Force       bool
	BackupMode  string // "separate" atau "combined"
	SuccessMsg  string
	LogPrefix   string
}

// BackupResult menyimpan hasil dari proses backup database.
type BackupResult struct {
	TotalDatabases      int
	SuccessfulBackups   int
	FailedBackups       int
	BackupInfo          []DatabaseBackupInfo
	FailedDatabases     map[string]string // map[databaseName]errorMessage
	FailedDatabaseInfos []FailedDatabaseInfo
	Errors              []string
	TotalTimeTaken      time.Duration
}

// FailedDatabaseInfo berisi informasi database yang gagal dibackup
type FailedDatabaseInfo struct {
	DatabaseName string `json:"database_name"`
	Error        string `json:"error"`
}

// BackupFileInfo menyimpan informasi ringkas tentang file backup.
type BackupFileInfo struct {
	Path         string
	ModTime      time.Time
	Size         int64
	DatabaseName string
}

// BackupMetadata menyimpan metadata lengkap untuk sebuah backup file
type BackupMetadata struct {
	BackupFile       string    `json:"backup_file"`                 // Path file backup
	BackupType       string    `json:"backup_type"`                 // "combined" atau "separated"
	DatabaseNames    []string  `json:"database_names"`              // List database yang di-backup
	Hostname         string    `json:"hostname"`                    // Database server hostname
	BackupStartTime  time.Time `json:"backup_start_time"`           // Waktu mulai backup
	BackupEndTime    time.Time `json:"backup_end_time"`             // Waktu selesai backup
	BackupDuration   string    `json:"backup_duration"`             // Duration dalam format human-readable
	FileSize         int64     `json:"file_size_bytes"`             // Ukuran file backup
	FileSizeHuman    string    `json:"file_size_human"`             // Ukuran file human-readable
	Compressed       bool      `json:"compressed"`                  // Apakah terkompresi
	CompressionType  string    `json:"compression_type,omitempty"`  // gzip, zstd, xz, dll
	Encrypted        bool      `json:"encrypted"`                   // Apakah terenkripsi
	MysqldumpVersion string    `json:"mysqldump_version,omitempty"` // Versi mysqldump
	MariaDBVersion   string    `json:"mariadb_version,omitempty"`   // Versi MariaDB/MySQL
	GTIDInfo         string    `json:"gtid_info,omitempty"`         // GTID information
	BackupStatus     string    `json:"backup_status"`               // "success", "partial", "failed"
	Warnings         []string  `json:"warnings,omitempty"`          // Warning messages
	GeneratedBy      string    `json:"generated_by"`                // Tool name dan version
	GeneratedAt      time.Time `json:"generated_at"`                // Waktu generate metadata
}

// BackupWriteResult menyimpan hasil dari operasi backup write
type BackupWriteResult struct {
	StderrOutput string // Output stderr dari mysqldump
	BytesWritten int64  // Total bytes written
}
