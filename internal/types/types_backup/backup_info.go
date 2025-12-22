// File : internal/types/types_backup/backup_info.go
// Deskripsi : Database backup info struct
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-22
// Last Modified : 2025-12-22

package types_backup

import "time"

// DatabaseBackupInfo berisi informasi database yang berhasil dibackup
type DatabaseBackupInfo struct {
	DatabaseName        string `json:"database_name"`
	OutputFile          string `json:"output_file"`
	FileSize            int64  `json:"file_size_bytes"`        // Ukuran file backup actual (compressed)
	FileSizeHuman       string `json:"file_size_human"`        // Ukuran file backup actual (human-readable)
	OriginalDBSize      int64  `json:"original_db_size_bytes"` // Ukuran database asli (sebelum backup)
	OriginalDBSizeHuman string `json:"original_db_size_human"` // Ukuran database asli (human-readable)
	Duration            string `json:"duration"`
	Status              string `json:"status"`                   // "success", "success_with_warnings", "failed"
	Warnings            string `json:"warnings,omitempty"`       // Warning/error messages dari mysqldump
	ErrorLogFile        string `json:"error_log_file,omitempty"` // Path ke file log error

	// Additional metadata (optional)
	BackupID       string    `json:"backup_id,omitempty"`
	StartTime      time.Time `json:"start_time,omitempty"`
	EndTime        time.Time `json:"end_time,omitempty"`
	ThroughputMBps float64   `json:"throughput_mb_per_sec,omitempty"`
	ManifestFile   string    `json:"manifest_file,omitempty"`
}
