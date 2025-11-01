package types

import "time"

// BackupFileInfo menyimpan informasi ringkas tentang file backup.
type BackupFileInfo struct {
	Path    string
	ModTime time.Time
	Size    int64
}

// BackupOptions menyimpan opsi konfigurasi untuk proses backup database.
type BackupOptions struct {
	Filter      FilterOptions
	Profile     ProfileInfo
	Compression CompressionOptions
	Encryption  EncryptionOptions
	Cleanup     CleanupOptions
	Background  bool
	CaptureGTID bool
	DryRun      bool
	OutputDir   string
}

// BackupConfig untuk konfigurasi backup entry point
type BackupEntryConfig struct {
	HeaderTitle string
	ShowOptions bool
	BackupMode  string // "separate" atau "combined"
	SuccessMsg  string
	LogPrefix   string
}
