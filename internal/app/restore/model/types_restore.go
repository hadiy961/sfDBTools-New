// File : internal/app/restore/model/types_restore.go
// Deskripsi : Type definitions untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 5 Januari 2026

package types

import "sfdbtools/internal/domain"

// RestoreSelectionEntry merepresentasikan satu baris dari CSV selection
type RestoreSelectionEntry struct {
	File       string // Path file backup
	DBName     string // Nama database target (boleh kosong → infer dari filename)
	EncKey     string // Kunci enkripsi file (wajib jika file terenkripsi)
	GrantsFile string // Path grants file (opsional)
}

// RestoreSelectionOptions opsi konfigurasi untuk restore selection
type RestoreSelectionOptions struct {
	Profile       domain.ProfileInfo    // Target profile
	CSV           string                // Path CSV input
	DropTarget    bool                  // Drop target db dahulu
	SkipBackup    bool                  // Skip backup pre-restore
	BackupOptions *RestoreBackupOptions // Opsi backup pre-restore
	Ticket        string                // Ticket number (wajib)
	Force         bool                  // Bypass confirmation
	DryRun        bool                  // Analisis saja
	StopOnError   bool                  // True = stop pada error pertama; False = lanjut (continue-on-error)
}

// RestoreSingleOptions menyimpan opsi konfigurasi untuk proses restore single database
type RestoreSingleOptions struct {
	Profile       domain.ProfileInfo    // Profile database target untuk restore
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
	Force         bool                  // Bypass confirmations / force mode
	StopOnError   bool                  // True = stop pada error pertama; False = lanjut (continue-on-error)
}

// RestoreBackupOptions opsi untuk backup sebelum restore
type RestoreBackupOptions struct {
	OutputDir   string // Direktori output untuk backup pre-restore (jika kosong, gunakan dari config)
	Compression domain.CompressionOptions
	Encryption  domain.EncryptionOptions
}

// RestorePrimaryOptions menyimpan opsi konfigurasi untuk proses restore primary database
type RestorePrimaryOptions struct {
	Profile            domain.ProfileInfo    // Profile database target untuk restore
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
	Force              bool                  // Bypass confirmations / force mode
	StopOnError        bool                  // True = stop pada error pertama; False = lanjut (continue-on-error)
}

// RestoreSecondaryOptions menyimpan opsi konfigurasi untuk proses restore secondary database.
// Secondary database mengikuti pola: dbsf_{nbc|biznet}_{client-code}_secondary_{instance}
// Sumber restore bisa berasal dari file backup atau dari backup database primary.
type RestoreSecondaryOptions struct {
	Profile       domain.ProfileInfo // Profile database target untuk restore
	DropTarget    bool               // Drop target database sebelum restore (default true)
	EncryptionKey string             // Kunci enkripsi untuk decrypt file backup (atau encrypt pre-backup primary)
	SkipBackup    bool               // Skip backup database target sebelum restore (default false)
	File          string             // Lokasi file backup yang akan di-restore (dipakai jika From=file)
	Ticket        string             // Ticket number untuk restore request (wajib)

	// Companion (dmart)
	IncludeDmart    bool   // Include companion database (_dmart) (default true)
	AutoDetectDmart bool   // Auto-detect file companion database (_dmart) (default true; only for From=file)
	CompanionFile   string // Lokasi file backup companion (_dmart) - optional, auto-detect jika kosong (From=file)

	// Secondary naming
	ClientCode string // Client code (akan menjadi dbsf_nbc_{client-code} / dbsf_biznet_{client-code})
	Instance   string // Instance secondary (suffix setelah _secondary_)

	// Source
	From      string // primary|file
	PrimaryDB string // Resolved primary database name (dipakai jika From=primary)

	// Target
	TargetDB      string                // Database secondary target untuk restore
	BackupOptions *RestoreBackupOptions // Opsi untuk backup sebelum restore (jika tidak skip)

	// Behavior
	DryRun      bool // Dry-run mode: validasi tanpa restore (default false)
	Force       bool // Bypass confirmations / force mode
	StopOnError bool // Reserved for consistency (default: stop)
}

// RestoreAllOptions opsi konfigurasi untuk restore all databases
type RestoreAllOptions struct {
	Profile       domain.ProfileInfo
	EncryptionKey string
	File          string
	Ticket        string
	GrantsFile    string // Optional: path file user grants (_users.sql)

	// Safety & Behavior
	BackupOptions *RestoreBackupOptions
	SkipBackup    bool
	DryRun        bool
	Force         bool
	StopOnError   bool
	DropTarget    bool

	// Filtering
	ExcludeDBs    []string // List DB yang akan di-skip
	SkipSystemDBs bool     // Skip mysql, sys, performance_schema, information_schema
	SkipGrants    bool     // Opsi tambahan jika ingin skip user grants (biasanya di DB mysql)
}

// RestoreCustomOptions opsi konfigurasi untuk restore custom (SFCola account detail → provision db+users+restore db+dmart)
// Catatan: field password disimpan hanya untuk eksekusi, jangan pernah di-log.
type RestoreCustomOptions struct {
	Profile       domain.ProfileInfo    // Target profile
	DropTarget    bool                  // Drop target db & dmart jika sudah ada
	SkipBackup    bool                  // Skip backup pre-restore
	BackupOptions *RestoreBackupOptions // Opsi backup pre-restore
	EncryptionKey string                // Kunci enkripsi untuk decrypt file backup
	Ticket        string                // Ticket number (wajib)
	Force         bool                  // Bypass confirmation
	DryRun        bool                  // Analisis saja
	StopOnError   bool                  // True = stop pada error pertama; False = lanjut (continue-on-error)

	// Extracted from pasted SFCola account detail
	Database      string
	DatabaseDmart string
	UserAdmin     string
	PassAdmin     string
	UserFin       string
	PassFin       string
	UserUser      string
	PassUser      string

	// Selected backup files
	DatabaseFile      string
	DatabaseDmartFile string
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
