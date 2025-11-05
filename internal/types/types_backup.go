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
	Mode        string         // "separate" atau "combined"
	ShowOptions bool           // Tampilkan opsi backup sebelum eksekusi
	NamePattern string         // Pola penamaan file backup
	File        BackupFileInfo // Nama file backup lengkap dengan ekstensi
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

// ChecksumInfo menyimpan informasi checksum untuk verifikasi backup
type ChecksumInfo struct {
	Algorithm     string    `json:"algorithm"`                // "sha256" atau "md5"
	Hash          string    `json:"hash"`                     // Hex-encoded hash value
	CalculatedAt  time.Time `json:"calculated_at"`            // Waktu kalkulasi checksum
	FileSize      int64     `json:"file_size"`                // Ukuran file saat checksum dihitung
	VerifiedAt    time.Time `json:"verified_at,omitempty"`    // Waktu verifikasi
	VerifyStatus  string    `json:"verify_status,omitempty"`  // "success", "failed", "skipped"
	VerifyMessage string    `json:"verify_message,omitempty"` // Pesan verifikasi
}

// BackupMetadata menyimpan metadata lengkap untuk sebuah backup file
type BackupMetadata struct {
	BackupFile       string         `json:"backup_file"`                 // Path file backup
	BackupType       string         `json:"backup_type"`                 // "combined" atau "separated"
	DatabaseNames    []string       `json:"database_names"`              // List database yang di-backup
	Hostname         string         `json:"hostname"`                    // Database server hostname
	BackupStartTime  time.Time      `json:"backup_start_time"`           // Waktu mulai backup
	BackupEndTime    time.Time      `json:"backup_end_time"`             // Waktu selesai backup
	BackupDuration   string         `json:"backup_duration"`             // Duration dalam format human-readable
	FileSize         int64          `json:"file_size_bytes"`             // Ukuran file backup
	FileSizeHuman    string         `json:"file_size_human"`             // Ukuran file human-readable
	Compressed       bool           `json:"compressed"`                  // Apakah terkompresi
	CompressionType  string         `json:"compression_type,omitempty"`  // gzip, zstd, xz, dll
	Encrypted        bool           `json:"encrypted"`                   // Apakah terenkripsi
	Checksums        []ChecksumInfo `json:"checksums"`                   // Multiple checksums (SHA256, MD5)
	MysqldumpVersion string         `json:"mysqldump_version,omitempty"` // Versi mysqldump
	MariaDBVersion   string         `json:"mariadb_version,omitempty"`   // Versi MariaDB/MySQL
	GTIDInfo         string         `json:"gtid_info,omitempty"`         // GTID information
	BackupStatus     string         `json:"backup_status"`               // "success", "partial", "failed"
	Warnings         []string       `json:"warnings,omitempty"`          // Warning messages
	GeneratedBy      string         `json:"generated_by"`                // Tool name dan version
	GeneratedAt      time.Time      `json:"generated_at"`                // Waktu generate metadata
}

// ChecksumResult menyimpan hasil verifikasi checksum
type ChecksumResult struct {
	FilePath     string        `json:"file_path"`
	SHA256       string        `json:"sha256"`
	MD5          string        `json:"md5"`
	Verified     bool          `json:"verified"`
	VerifyTime   time.Duration `json:"verify_time"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// BackupWriteResult menyimpan hasil dari operasi backup write
type BackupWriteResult struct {
	StderrOutput string // Output stderr dari mysqldump
	SHA256Hash   string // SHA256 checksum dari backup file
	MD5Hash      string // MD5 checksum dari backup file
	BytesWritten int64  // Total bytes written
}

// ChecksumVerificationResult menyimpan hasil verifikasi checksum
type ChecksumVerificationResult struct {
	FilePath         string `json:"file_path"`
	ExpectedSHA256   string `json:"expected_sha256"`
	ExpectedMD5      string `json:"expected_md5"`
	CalculatedSHA256 string `json:"calculated_sha256"`
	CalculatedMD5    string `json:"calculated_md5"`
	SHA256Match      bool   `json:"sha256_match"`
	MD5Match         bool   `json:"md5_match"`
	Success          bool   `json:"success"`
	Error            string `json:"error,omitempty"`
}
