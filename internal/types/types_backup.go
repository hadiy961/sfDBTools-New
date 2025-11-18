package types

import (
	"encoding/json"
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
	CaptureGTID bool // Tangkap informasi GTID saat backup (hanya untuk combined)
	ExcludeUser bool // Exclude user grants dari export (default: false = export user)
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

// MarshalJSON customizes JSON output so BackupStartTime and BackupEndTime are
// formatted as MariaDB-compatible DATETIME strings: "2006-01-02 15:04:05".
func (b BackupMetadata) MarshalJSON() ([]byte, error) {
	type metaJSON struct {
		BackupFile       string    `json:"backup_file"`
		BackupType       string    `json:"backup_type"`
		DatabaseNames    []string  `json:"database_names"`
		Hostname         string    `json:"hostname"`
		BackupStartTime  string    `json:"backup_start_time"`
		BackupEndTime    string    `json:"backup_end_time"`
		BackupDuration   string    `json:"backup_duration"`
		FileSize         int64     `json:"file_size_bytes"`
		FileSizeHuman    string    `json:"file_size_human"`
		Compressed       bool      `json:"compressed"`
		CompressionType  string    `json:"compression_type,omitempty"`
		Encrypted        bool      `json:"encrypted"`
		MysqldumpVersion string    `json:"mysqldump_version,omitempty"`
		MariaDBVersion   string    `json:"mariadb_version,omitempty"`
		GTIDInfo         string    `json:"gtid_info,omitempty"`
		BackupStatus     string    `json:"backup_status"`
		Warnings         []string  `json:"warnings,omitempty"`
		GeneratedBy      string    `json:"generated_by"`
		GeneratedAt      time.Time `json:"generated_at"`
	}

	mj := metaJSON{
		BackupFile:       b.BackupFile,
		BackupType:       b.BackupType,
		DatabaseNames:    b.DatabaseNames,
		Hostname:         b.Hostname,
		BackupStartTime:  b.BackupStartTime.Format("2006-01-02 15:04:05"),
		BackupEndTime:    b.BackupEndTime.Format("2006-01-02 15:04:05"),
		BackupDuration:   b.BackupDuration,
		FileSize:         b.FileSize,
		FileSizeHuman:    b.FileSizeHuman,
		Compressed:       b.Compressed,
		CompressionType:  b.CompressionType,
		Encrypted:        b.Encrypted,
		MysqldumpVersion: b.MysqldumpVersion,
		MariaDBVersion:   b.MariaDBVersion,
		GTIDInfo:         b.GTIDInfo,
		BackupStatus:     b.BackupStatus,
		Warnings:         b.Warnings,
		GeneratedBy:      b.GeneratedBy,
		GeneratedAt:      b.GeneratedAt,
	}

	return json.MarshalIndent(mj, "", "  ")
}

// UnmarshalJSON accepts either MariaDB-compatible DATETIME strings
// ("2006-01-02 15:04:05") or RFC3339 timestamps for the start/end fields.
func (b *BackupMetadata) UnmarshalJSON(data []byte) error {
	type metaJSONIn struct {
		BackupFile       string    `json:"backup_file"`
		BackupType       string    `json:"backup_type"`
		DatabaseNames    []string  `json:"database_names"`
		Hostname         string    `json:"hostname"`
		BackupStartTime  string    `json:"backup_start_time"`
		BackupEndTime    string    `json:"backup_end_time"`
		BackupDuration   string    `json:"backup_duration"`
		FileSize         int64     `json:"file_size_bytes"`
		FileSizeHuman    string    `json:"file_size_human"`
		Compressed       bool      `json:"compressed"`
		CompressionType  string    `json:"compression_type,omitempty"`
		Encrypted        bool      `json:"encrypted"`
		MysqldumpVersion string    `json:"mysqldump_version,omitempty"`
		MariaDBVersion   string    `json:"mariadb_version,omitempty"`
		GTIDInfo         string    `json:"gtid_info,omitempty"`
		BackupStatus     string    `json:"backup_status"`
		Warnings         []string  `json:"warnings,omitempty"`
		GeneratedBy      string    `json:"generated_by"`
		GeneratedAt      time.Time `json:"generated_at"`
	}

	var mj metaJSONIn
	if err := json.Unmarshal(data, &mj); err != nil {
		return err
	}

	// parse start time - accept MariaDB format first, then RFC3339 as fallback
	var st, et time.Time
	var pErr error
	if mj.BackupStartTime != "" {
		st, pErr = time.ParseInLocation("2006-01-02 15:04:05", mj.BackupStartTime, time.Local)
		if pErr != nil {
			st, pErr = time.Parse(time.RFC3339, mj.BackupStartTime)
			if pErr != nil {
				return pErr
			}
		}
	}
	if mj.BackupEndTime != "" {
		et, pErr = time.ParseInLocation("2006-01-02 15:04:05", mj.BackupEndTime, time.Local)
		if pErr != nil {
			et, pErr = time.Parse(time.RFC3339, mj.BackupEndTime)
			if pErr != nil {
				return pErr
			}
		}
	}

	b.BackupFile = mj.BackupFile
	b.BackupType = mj.BackupType
	b.DatabaseNames = mj.DatabaseNames
	b.Hostname = mj.Hostname
	b.BackupStartTime = st
	b.BackupEndTime = et
	b.BackupDuration = mj.BackupDuration
	b.FileSize = mj.FileSize
	b.FileSizeHuman = mj.FileSizeHuman
	b.Compressed = mj.Compressed
	b.CompressionType = mj.CompressionType
	b.Encrypted = mj.Encrypted
	b.MysqldumpVersion = mj.MysqldumpVersion
	b.MariaDBVersion = mj.MariaDBVersion
	b.GTIDInfo = mj.GTIDInfo
	b.BackupStatus = mj.BackupStatus
	b.Warnings = mj.Warnings
	b.GeneratedBy = mj.GeneratedBy
	b.GeneratedAt = mj.GeneratedAt

	return nil
}

// BackupWriteResult menyimpan hasil dari operasi backup write
type BackupWriteResult struct {
	StderrOutput string // Output stderr dari mysqldump
	BytesWritten int64  // Total bytes written
}
