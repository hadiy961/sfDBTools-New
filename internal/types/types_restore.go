// File : internal/types/types_restore.go
// Deskripsi : Type definitions untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-16

package types

// RestoreSingleOptions menyimpan opsi konfigurasi untuk proses restore single database
type RestoreSingleOptions struct {
	Profile       ProfileInfo           // Profile database target untuk restore
	DropTarget    bool                  // Drop target database sebelum restore (default true)
	EncryptionKey string                // Kunci enkripsi untuk decrypt file backup
	SkipBackup    bool                  // Skip backup sebelum restore (default false)
	File          string                // Lokasi file backup yang akan di-restore
	Ticket        string                // Ticket number untuk restore request (wajib)
	TargetDB      string                // Database target untuk restore
	BackupOptions *RestoreBackupOptions // Opsi untuk backup sebelum restore (jika tidak skip)
	GrantsFile    string                // Lokasi file user grants (optional, jika ada)
	SkipGrants    bool                  // Skip restore user grants (default false)
	DryRun        bool                  // Dry-run mode: validasi tanpa restore (default false)
}

// RestoreBackupOptions opsi untuk backup sebelum restore
type RestoreBackupOptions struct {
	OutputDir   string // Direktori output untuk backup pre-restore (jika kosong, gunakan dari config)
	Compression CompressionOptions
	Encryption  EncryptionOptions
}

// CompressionOptions menyimpan opsi kompresi
type CompressionOptions struct {
	Enabled bool
	Type    string
	Level   int
}

// EncryptionOptions menyimpan opsi enkripsi
type EncryptionOptions struct {
	Enabled bool
	Key     string
}

// RestorePrimaryOptions menyimpan opsi konfigurasi untuk proses restore primary database
type RestorePrimaryOptions struct {
	Profile            ProfileInfo           // Profile database target untuk restore
	DropTarget         bool                  // Drop target database sebelum restore (default true)
	EncryptionKey      string                // Kunci enkripsi untuk decrypt file backup
	SkipBackup         bool                  // Skip backup sebelum restore (default false)
	File               string                // Lokasi file backup primary yang akan di-restore
	CompanionFile      string                // Lokasi file backup companion (dmart) - optional, auto-detect jika kosong
	Ticket             string                // Ticket number untuk restore request (wajib)
	TargetDB           string                // Database primary target untuk restore
	BackupOptions      *RestoreBackupOptions // Opsi untuk backup sebelum restore (jika tidak skip)
	GrantsFile         string                // Lokasi file user grants (optional, jika ada)
	SkipGrants         bool                  // Skip restore user grants (default false)
	IncludeDmart       bool                  // Include companion database _dmart (default true)
	AutoDetectDmart    bool                  // Auto-detect file companion database _dmart (default true)
	ConfirmIfNotExists bool                  // Konfirmasi jika database belum ada (default true)
	DryRun             bool                  // Dry-run mode: validasi tanpa restore (default false)
}

// RestoreAllOptions opsi konfigurasi untuk restore all databases
type RestoreAllOptions struct {
	Profile       ProfileInfo
	EncryptionKey string
	File          string
	Ticket        string

	// Safety & Behavior
	BackupOptions *RestoreBackupOptions
	SkipBackup    bool
	DryRun        bool
	Force         bool
	StopOnError   bool
	DropTarget    bool // Drop semua database non-sistem sebelum restore

	// Filtering
	ExcludeDBs    []string // List DB yang akan di-skip
	SkipSystemDBs bool     // Skip mysql, sys, performance_schema, information_schema
	SkipGrants    bool     // Opsi tambahan jika ingin skip user grants (biasanya di DB mysql)
}

// RestoreResult menyimpan hasil restore operation
type RestoreResult struct {
	Success          bool
	TargetDB         string
	SourceFile       string
	CompanionDB      string // Companion database yang di-restore (biasanya _dmart)
	CompanionFile    string // File companion yang di-restore
	BackupFile       string // File backup pre-restore (jika ada)
	CompanionBackup  string // File backup companion pre-restore (jika ada)
	DroppedDB        bool   // Apakah database di-drop sebelum restore
	DroppedCompanion bool   // Apakah companion database di-drop sebelum restore
	GrantsFile       string // File user grants yang di-restore (jika ada)
	GrantsRestored   bool   // Apakah user grants berhasil di-restore
	Error            error
	Duration         string
}
