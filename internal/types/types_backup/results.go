// File : internal/types/types_backup/results.go
// Deskripsi : Result structs untuk backup operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-08

package types_backup

import (
	"encoding/json"
	"sfDBTools/internal/types"
	"time"
)

// BackupResult menyimpan hasil dari proses backup database.
type BackupResult struct {
	TotalDatabases      int
	SuccessfulBackups   int
	FailedBackups       int
	BackupInfo          []types.DatabaseBackupInfo
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

// BackupWriteResult menyimpan hasil dari operasi backup write
type BackupWriteResult struct {
	StderrOutput string // Output stderr dari mysqldump
	BytesWritten int64  // Total bytes written
	FileSize     int64  // File size after write (sama dengan BytesWritten untuk consistency)
}

// BackupMetadata menyimpan metadata lengkap untuk sebuah backup file
type BackupMetadata struct {
	BackupFile        string                 `json:"backup_file"`                  // Path file backup
	BackupType        string                 `json:"backup_type"`                  // "combined" atau "separated"
	DatabaseNames     []string               `json:"database_names"`               // List database yang di-backup
	ExcludedDatabases []string               `json:"excluded_databases,omitempty"` // List database yang dikecualikan (untuk mode 'all')
	DatabaseDetails   []DatabaseBackupDetail `json:"database_details,omitempty"`   // Detail per database untuk primary/secondary
	Hostname          string                 `json:"hostname"`                     // Database server hostname
	BackupStartTime   time.Time              `json:"backup_start_time"`            // Waktu mulai backup
	BackupEndTime     time.Time              `json:"backup_end_time"`              // Waktu selesai backup
	BackupDuration    string                 `json:"backup_duration"`              // Duration dalam format human-readable
	FileSize          int64                  `json:"file_size_bytes"`              // Ukuran file backup
	FileSizeHuman     string                 `json:"file_size_human"`              // Ukuran file human-readable
	Compressed        bool                   `json:"compressed"`                   // Apakah terkompresi
	CompressionType   string                 `json:"compression_type,omitempty"`   // gzip, zstd, xz, dll
	Encrypted         bool                   `json:"encrypted"`                    // Apakah terenkripsi
	MysqldumpVersion  string                 `json:"mysqldump_version,omitempty"`  // Versi mysqldump
	MariaDBVersion    string                 `json:"mariadb_version,omitempty"`    // Versi MariaDB/MySQL
	GTIDInfo          string                 `json:"gtid_info,omitempty"`          // GTID information
	GTIDFile          string                 `json:"gtid_file,omitempty"`          // Path ke file GTID
	UserGrantsFile    string                 `json:"user_grants_file,omitempty"`   // Path ke file user grants
	BackupStatus      string                 `json:"backup_status"`                // "success", "partial", "failed"
	Warnings          []string               `json:"warnings,omitempty"`           // Warning messages
	GeneratedBy       string                 `json:"generated_by"`                 // Tool name dan version
	GeneratedAt       time.Time              `json:"generated_at"`                 // Waktu generate metadata
	Ticket            string                 `json:"ticket"`                       // Ticket number untuk request backup
	// Replication information
	ReplicationUser     string `json:"replication_user,omitempty"`     // User replikasi
	ReplicationPassword string `json:"replication_password,omitempty"` // Password replikasi
	SourceHost          string `json:"source_host,omitempty"`          // IP/Host sumber database
	SourcePort          int    `json:"source_port,omitempty"`          // Port sumber database
}

// DatabaseBackupDetail menyimpan detail backup per database (untuk primary/secondary mode)
type DatabaseBackupDetail struct {
	DatabaseName  string `json:"database_name"`
	BackupFile    string `json:"backup_file"`
	FileSizeBytes int64  `json:"file_size_bytes"`
	FileSizeHuman string `json:"file_size_human"`
}

// MarshalJSON customizes JSON output dengan struktur yang terorganisir dalam grup
// untuk memudahkan pembacaan dan pemahaman metadata backup
func (b BackupMetadata) MarshalJSON() ([]byte, error) {
	// Grup untuk informasi file backup
	type backupInfo struct {
		File              string   `json:"file"`
		Type              string   `json:"type"`
		Status            string   `json:"status"`
		Databases         []string `json:"databases"`
		ExcludedDatabases []string `json:"excluded_databases"` // Hapus omitempty untuk testing
	}

	// Grup untuk informasi waktu
	type timeInfo struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Duration  string `json:"duration"`
	}

	// Grup untuk informasi file
	type fileInfo struct {
		SizeBytes int64  `json:"size_bytes"`
		SizeHuman string `json:"size_human"`
	}

	// Grup untuk informasi kompresi
	type compressionInfo struct {
		Enabled bool   `json:"enabled"`
		Type    string `json:"type,omitempty"`
	}

	// Grup untuk informasi enkripsi
	type encryptionInfo struct {
		Enabled bool `json:"enabled"`
	}

	// Grup untuk informasi sumber database
	type sourceInfo struct {
		Hostname string `json:"hostname"`
		Host     string `json:"host,omitempty"`
		Port     int    `json:"port,omitempty"`
	}

	// Grup untuk informasi replikasi
	type replicationInfo struct {
		User     string `json:"user,omitempty"`
		Password string `json:"password,omitempty"`
		GTIDInfo string `json:"gtid_info,omitempty"`
	}

	// Grup untuk informasi versi
	type versionInfo struct {
		MysqldumpVersion string `json:"mysqldump,omitempty"`
		MariaDBVersion   string `json:"mariadb,omitempty"`
	}

	// Grup untuk file tambahan
	type additionalFiles struct {
		UserGrants string `json:"user_grants,omitempty"`
	}

	// Grup untuk informasi generator
	type generatorInfo struct {
		GeneratedBy string    `json:"generated_by"`
		GeneratedAt time.Time `json:"generated_at"`
	}

	// Struct utama dengan grouping
	metaJSON := struct {
		Backup          backupInfo             `json:"backup"`
		DatabaseDetails []DatabaseBackupDetail `json:"database_details,omitempty"`
		Time            timeInfo               `json:"time"`
		File            fileInfo               `json:"file"`
		Compression     compressionInfo        `json:"compression"`
		Encryption      encryptionInfo         `json:"encryption"`
		Source          sourceInfo             `json:"source"`
		Replication     replicationInfo        `json:"replication"`
		Version         versionInfo            `json:"version,omitempty"`
		Additional      additionalFiles        `json:"additional_files,omitempty"`
		Generator       generatorInfo          `json:"generator"`
		Warnings        []string               `json:"warnings,omitempty"`
	}{
		Backup: backupInfo{
			File:              b.BackupFile,
			Type:              b.BackupType,
			Status:            b.BackupStatus,
			Databases:         b.DatabaseNames,
			ExcludedDatabases: b.ExcludedDatabases,
		},
		DatabaseDetails: b.DatabaseDetails,
		Time: timeInfo{
			StartTime: b.BackupStartTime.Format("2006-01-02 15:04:05"),
			EndTime:   b.BackupEndTime.Format("2006-01-02 15:04:05"),
			Duration:  b.BackupDuration,
		},
		File: fileInfo{
			SizeBytes: b.FileSize,
			SizeHuman: b.FileSizeHuman,
		},
		Compression: compressionInfo{
			Enabled: b.Compressed,
			Type:    b.CompressionType,
		},
		Encryption: encryptionInfo{
			Enabled: b.Encrypted,
		},
		Source: sourceInfo{
			Hostname: b.Hostname,
			Host:     b.SourceHost,
			Port:     b.SourcePort,
		},
		Replication: replicationInfo{
			User:     b.ReplicationUser,
			Password: b.ReplicationPassword,
			GTIDInfo: b.GTIDInfo,
		},
		Version: versionInfo{
			MysqldumpVersion: b.MysqldumpVersion,
			MariaDBVersion:   b.MariaDBVersion,
		},
		Additional: additionalFiles{
			UserGrants: b.UserGrantsFile,
		},
		Generator: generatorInfo{
			GeneratedBy: b.GeneratedBy,
			GeneratedAt: b.GeneratedAt,
		},
		Warnings: b.Warnings,
	}

	return json.MarshalIndent(metaJSON, "", "  ")
}

// UnmarshalJSON accepts either MariaDB-compatible DATETIME strings
// ("2006-01-02 15:04:05") or RFC3339 timestamps for the start/end fields.
// Mendukung struktur flat (backward compatibility) dan struktur grup (format baru)
func (b *BackupMetadata) UnmarshalJSON(data []byte) error {
	// Coba parse dengan struktur grup dulu
	type backupInfo struct {
		File              string   `json:"file"`
		Type              string   `json:"type"`
		Status            string   `json:"status"`
		Databases         []string `json:"databases"`
		ExcludedDatabases []string `json:"excluded_databases"`
	}
	type timeInfo struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Duration  string `json:"duration"`
	}
	type fileInfo struct {
		SizeBytes int64  `json:"size_bytes"`
		SizeHuman string `json:"size_human"`
	}
	type compressionInfo struct {
		Enabled bool   `json:"enabled"`
		Type    string `json:"type,omitempty"`
	}
	type encryptionInfo struct {
		Enabled bool `json:"enabled"`
	}
	type sourceInfo struct {
		Hostname string `json:"hostname"`
		Host     string `json:"host,omitempty"`
		Port     int    `json:"port,omitempty"`
	}
	type replicationInfo struct {
		User     string `json:"user,omitempty"`
		Password string `json:"password,omitempty"`
		GTIDInfo string `json:"gtid_info,omitempty"`
	}
	type versionInfo struct {
		MysqldumpVersion string `json:"mysqldump,omitempty"`
		MariaDBVersion   string `json:"mariadb,omitempty"`
	}
	type additionalFiles struct {
		UserGrants string `json:"user_grants,omitempty"`
	}
	type generatorInfo struct {
		GeneratedBy string    `json:"generated_by"`
		GeneratedAt time.Time `json:"generated_at"`
	}

	type metaJSONGrouped struct {
		Backup          backupInfo             `json:"backup"`
		DatabaseDetails []DatabaseBackupDetail `json:"database_details,omitempty"`
		Time            timeInfo               `json:"time"`
		File            fileInfo               `json:"file"`
		Compression     compressionInfo        `json:"compression"`
		Encryption      encryptionInfo         `json:"encryption"`
		Source          sourceInfo             `json:"source"`
		Replication     replicationInfo        `json:"replication"`
		Version         versionInfo            `json:"version,omitempty"`
		Additional      additionalFiles        `json:"additional_files,omitempty"`
		Generator       generatorInfo          `json:"generator"`
		Warnings        []string               `json:"warnings,omitempty"`
	}

	var grouped metaJSONGrouped
	if err := json.Unmarshal(data, &grouped); err == nil && grouped.Backup.File != "" {
		// Parse grouped structure
		var st, et time.Time
		var pErr error
		if grouped.Time.StartTime != "" {
			st, pErr = time.ParseInLocation("2006-01-02 15:04:05", grouped.Time.StartTime, time.Local)
			if pErr != nil {
				st, pErr = time.Parse(time.RFC3339, grouped.Time.StartTime)
				if pErr != nil {
					return pErr
				}
			}
		}
		if grouped.Time.EndTime != "" {
			et, pErr = time.ParseInLocation("2006-01-02 15:04:05", grouped.Time.EndTime, time.Local)
			if pErr != nil {
				et, pErr = time.Parse(time.RFC3339, grouped.Time.EndTime)
				if pErr != nil {
					return pErr
				}
			}
		}

		b.BackupFile = grouped.Backup.File
		b.BackupType = grouped.Backup.Type
		b.BackupStatus = grouped.Backup.Status
		b.DatabaseNames = grouped.Backup.Databases
		b.ExcludedDatabases = grouped.Backup.ExcludedDatabases
		b.DatabaseDetails = grouped.DatabaseDetails
		b.Hostname = grouped.Source.Hostname
		b.SourceHost = grouped.Source.Host
		b.SourcePort = grouped.Source.Port
		b.BackupStartTime = st
		b.BackupEndTime = et
		b.BackupDuration = grouped.Time.Duration
		b.FileSize = grouped.File.SizeBytes
		b.FileSizeHuman = grouped.File.SizeHuman
		b.Compressed = grouped.Compression.Enabled
		b.CompressionType = grouped.Compression.Type
		b.Encrypted = grouped.Encryption.Enabled
		b.MysqldumpVersion = grouped.Version.MysqldumpVersion
		b.MariaDBVersion = grouped.Version.MariaDBVersion
		b.GTIDInfo = grouped.Replication.GTIDInfo
		b.ReplicationUser = grouped.Replication.User
		b.ReplicationPassword = grouped.Replication.Password
		b.UserGrantsFile = grouped.Additional.UserGrants
		b.GeneratedBy = grouped.Generator.GeneratedBy
		b.GeneratedAt = grouped.Generator.GeneratedAt
		b.Warnings = grouped.Warnings

		return nil
	}

	// Fallback: parse dengan struktur flat untuk backward compatibility
	type metaJSONIn struct {
		BackupFile          string    `json:"backup_file"`
		BackupType          string    `json:"backup_type"`
		DatabaseNames       []string  `json:"database_names"`
		Hostname            string    `json:"hostname"`
		BackupStartTime     string    `json:"backup_start_time"`
		BackupEndTime       string    `json:"backup_end_time"`
		BackupDuration      string    `json:"backup_duration"`
		FileSize            int64     `json:"file_size_bytes"`
		FileSizeHuman       string    `json:"file_size_human"`
		Compressed          bool      `json:"compressed"`
		CompressionType     string    `json:"compression_type,omitempty"`
		Encrypted           bool      `json:"encrypted"`
		MysqldumpVersion    string    `json:"mysqldump_version,omitempty"`
		MariaDBVersion      string    `json:"mariadb_version,omitempty"`
		GTIDInfo            string    `json:"gtid_info,omitempty"`
		BackupStatus        string    `json:"backup_status"`
		Warnings            []string  `json:"warnings,omitempty"`
		GeneratedBy         string    `json:"generated_by"`
		GeneratedAt         time.Time `json:"generated_at"`
		ReplicationUser     string    `json:"replication_user,omitempty"`
		ReplicationPassword string    `json:"replication_password,omitempty"`
		SourceHost          string    `json:"source_host,omitempty"`
		SourcePort          int       `json:"source_port,omitempty"`
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
	b.ReplicationUser = mj.ReplicationUser
	b.ReplicationPassword = mj.ReplicationPassword
	b.SourceHost = mj.SourceHost
	b.SourcePort = mj.SourcePort

	return nil
}
