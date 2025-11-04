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
	Background  bool
	DryRun      bool
	OutputDir   string
	Mode        string // "separate" atau "combined"
	ShowOptions bool   // Tampilkan opsi backup sebelum eksekusi
	NamePattern string // Pola penamaan file backup
	Entry       BackupEntryConfig
	CaptureGTID bool // Tangkap informasi GTID saat backup
}

// BackupConfig untuk konfigurasi backup entry point
type BackupEntryConfig struct {
	HeaderTitle string
	ShowOptions bool
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
	Path    string
	ModTime time.Time
	Size    int64
}
