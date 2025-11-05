// File : internal/types/types_restore.go
// Deskripsi : Type definitions untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package types

import "time"

// RestoreOptions menyimpan opsi konfigurasi untuk proses restore database
type RestoreOptions struct {
	// Source backup file
	SourceFile string
	
	// Target profile and authentication
	TargetProfile    string
	TargetProfileKey string
	
	// Encryption key untuk decrypt backup file
	EncryptionKey string
	
	// Target database name (untuk single restore)
	// Jika kosong, gunakan nama database dari backup file
	TargetDB string
	
	// Mode restore: "single", "all", "multi"
	Mode string
	
	// Verify checksum sebelum restore
	VerifyChecksum bool
	
	// Force restore (skip confirmation)
	Force bool
	
	// Dry run (simulate without actual restore)
	DryRun bool
	
	// Show options before execution
	ShowOptions bool
}

// RestoreResult menyimpan hasil dari proses restore database
type RestoreResult struct {
	TotalDatabases    int
	SuccessfulRestore int
	FailedRestore     int
	RestoreInfo       []DatabaseRestoreInfo
	FailedDatabases   map[string]string // map[databaseName]errorMessage
	Errors            []string
	TotalTimeTaken    time.Duration
	VerificationInfo  *RestoreVerificationInfo
}

// DatabaseRestoreInfo menyimpan informasi detail tentang restore satu database
type DatabaseRestoreInfo struct {
	DatabaseName   string        `json:"database_name"`
	SourceFile     string        `json:"source_file"`
	TargetDatabase string        `json:"target_database"`
	FileSize       int64         `json:"file_size_bytes"`
	FileSizeHuman  string        `json:"file_size_human"`
	Duration       string        `json:"duration"`
	Status         string        `json:"status"` // "success", "failed", "skipped"
	Warnings       string        `json:"warnings,omitempty"`
	ErrorMessage   string        `json:"error_message,omitempty"`
	Verified       bool          `json:"verified"`
}

// RestoreVerificationInfo menyimpan informasi verifikasi backup sebelum restore
type RestoreVerificationInfo struct {
	BackupFile       string    `json:"backup_file"`
	FileSize         int64     `json:"file_size_bytes"`
	Encrypted        bool      `json:"encrypted"`
	Compressed       bool      `json:"compressed"`
	CompressionType  string    `json:"compression_type,omitempty"`
	ExpectedSHA256   string    `json:"expected_sha256,omitempty"`
	ExpectedMD5      string    `json:"expected_md5,omitempty"`
	CalculatedSHA256 string    `json:"calculated_sha256,omitempty"`
	CalculatedMD5    string    `json:"calculated_md5,omitempty"`
	ChecksumMatch    bool      `json:"checksum_match"`
	VerificationTime time.Time `json:"verification_time"`
	ErrorMessage     string    `json:"error_message,omitempty"`
}

// RestoreEntryConfig untuk konfigurasi restore entry point
type RestoreEntryConfig struct {
	HeaderTitle string
	ShowOptions bool
	RestoreMode string // "single", "all", "multi"
	SuccessMsg  string
	LogPrefix   string
}
