// File : internal/backup/metadata/builder.go
// Deskripsi : Builder pattern for backup metadata construction
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-22
// Last Modified : 2025-12-22

package metadata

import (
	"fmt"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/shared/global"
	"time"
)

// DatabaseBackupInfoBuilder membantu construct DatabaseBackupInfo dengan konsisten
type DatabaseBackupInfoBuilder struct {
	DatabaseName string
	OutputFile   string
	FileSize     int64
	Duration     time.Duration
	Status       string
	Warnings     string
	StartTime    time.Time
	EndTime      time.Time
	ManifestFile string
}

// Build membuat DatabaseBackupInfo dari builder
func (b *DatabaseBackupInfoBuilder) Build() types_backup.DatabaseBackupInfo {
	// Calculate throughput
	var throughputMBs float64
	if b.Duration.Seconds() > 0 {
		throughputMBs = float64(b.FileSize) / (1024.0 * 1024.0) / b.Duration.Seconds()
	}

	// Generate backup ID
	backupID := fmt.Sprintf("bk-%d", time.Now().UnixNano())

	return types_backup.DatabaseBackupInfo{
		DatabaseName:   b.DatabaseName,
		OutputFile:     b.OutputFile,
		FileSize:       b.FileSize,
		FileSizeHuman:  global.FormatFileSize(b.FileSize),
		Duration:       global.FormatDuration(b.Duration),
		Status:         b.Status,
		Warnings:       b.Warnings,
		BackupID:       backupID,
		StartTime:      b.StartTime,
		EndTime:        b.EndTime,
		ThroughputMBps: throughputMBs,
		ManifestFile:   b.ManifestFile,
	}
}
