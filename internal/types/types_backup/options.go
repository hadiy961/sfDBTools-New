// File : internal/types/types_backup/options.go
// Deskripsi : Options dan config structs untuk backup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-08

package types_backup

import (
	"sfDBTools/internal/types"
	"time"
)

// BackupExecutionConfig konfigurasi untuk backup execution
type BackupExecutionConfig struct {
	DBName       string
	DBList       []string
	OutputPath   string
	BackupType   string
	TotalDBFound int
	IsMultiDB    bool
	ProgressChan chan<- string `json:"-"`
}

// BackupDBOptions menyimpan opsi konfigurasi untuk proses backup database.
type BackupDBOptions struct {
	Filter          types.FilterOptions
	Profile         types.ProfileInfo
	Compression     types.CompressionOptions
	Encryption      types.EncryptionOptions
	Cleanup         types.CleanupOptions
	DryRun          bool
	OutputDir       string
	Mode            string         // "separate" atau "combined"
	Force           bool           // Tampilkan opsi backup sebelum eksekusi
	File            BackupFileInfo // Nama file backup lengkap dengan ekstensi
	Entry           BackupEntryConfig
	CaptureGTID     bool            // Tangkap informasi GTID saat backup (hanya untuk combined)
	ExcludeUser     bool            // Exclude user grants dari export (default: false = export user)
	DBName          string          // Nama database untuk backup single/primary/secondary
	IncludeDmart    bool            // Sertakan database <db>_dmart jika tersedia (hanya primary/secondary)
	CompanionStatus map[string]bool `json:"-"` // Status ketersediaan companion db (hanya primary/secondary)
	ClientCode      string          // Client code untuk filter database (primary/secondary)
	Instance        string          // Instance name untuk filter secondary database
	Ticket          string          // Ticket number untuk request backup (wajib)
}

// BackupEntryConfig untuk konfigurasi backup entry point
type BackupEntryConfig struct {
	HeaderTitle string
	Force       bool
	BackupMode  string // "separate" atau "combined"
	SuccessMsg  string
	LogPrefix   string
}

// BackupFileInfo menyimpan informasi ringkas tentang file backup.
type BackupFileInfo struct {
	Path         string
	ModTime      time.Time
	Size         int64
	DatabaseName string
	Filename     string
}
