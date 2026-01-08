// File : internal/backup/metadata/generator.go
// Deskripsi : Metadata generation untuk backup operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-22
// Last Modified : 2025-12-22

package metadata

import (
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/global"
	"time"
)

// GenerateBackupMetadata membuat metadata backup dari config
func GenerateBackupMetadata(cfg types_backup.MetadataConfig) *types_backup.BackupMetadata {
	meta := &types_backup.BackupMetadata{
		BackupFile:        cfg.BackupFile,
		BackupType:        cfg.BackupType,
		DatabaseNames:     cfg.DatabaseNames,
		ExcludedDatabases: cfg.ExcludedDatabases,
		Hostname:          cfg.Hostname,
		BackupStartTime:   cfg.StartTime,
		BackupEndTime:     cfg.EndTime,
		BackupDuration:    global.FormatDuration(cfg.Duration),
		FileSize:          cfg.FileSize,
		FileSizeHuman:     global.FormatFileSize(cfg.FileSize),
		Compressed:        cfg.Compressed,
		CompressionType:   cfg.CompressionType,
		Encrypted:         cfg.Encrypted,
		ExcludeData:       cfg.ExcludeData,
		BackupStatus:      cfg.BackupStatus,
		Warnings:          cfg.Warnings,
		GeneratedBy:       "sfdbtools",
		GeneratedAt:       time.Now(),
		Ticket:            cfg.Ticket,
		// Additional files
		UserGrantsFile: cfg.UserGrantsFile,
		// Version information
		MysqldumpVersion: cfg.MysqldumpVersion,
		MariaDBVersion:   cfg.MariaDBVersion,
	}

	// GTID info dan replication info hanya disimpan untuk mode non-separated (combined, single, dll)
	// Untuk mode separated/multi-file, informasi ini tidak relevan karena setiap database dibackup terpisah
	if cfg.BackupType != consts.ModeSeparated {
		meta.GTIDInfo = cfg.GTIDInfo
		meta.ReplicationUser = cfg.ReplicationUser
		meta.ReplicationPassword = cfg.ReplicationPassword
		meta.SourceHost = cfg.SourceHost
		meta.SourcePort = cfg.SourcePort
	}

	if cfg.StderrOutput != "" {
		meta.Warnings = append(meta.Warnings, cfg.StderrOutput)
	}

	return meta
}
