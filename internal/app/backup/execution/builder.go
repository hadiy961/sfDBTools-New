// File : internal/backup/execution/builder.go
// Deskripsi : Builder functions untuk DatabaseBackupInfo dan metadata generation
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-31
// Last Modified : 2026-01-02

package execution

import (
	"strings"
	"time"

	"sfdbtools/internal/app/backup/metadata"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/timex"
)

// buildDryRunInfo membuat DatabaseBackupInfo untuk dry-run mode.
func (e *Engine) buildDryRunInfo(
	cfg types_backup.BackupExecutionConfig,
	args []string,
	timer *timex.Timer,
	startTime time.Time,
) types_backup.DatabaseBackupInfo {
	if cfg.IsMultiDB {
		e.Log.Info("[DRY-RUN] Akan backup database: " + strings.Join(cfg.DBList, ", "))
	} else {
		e.Log.Infof("[DRY-RUN] Akan backup database: %s", cfg.DBName)
	}
	e.Log.Info("[DRY-RUN] Output file: " + cfg.OutputPath)
	e.Log.Debug("[DRY-RUN] Mysqldump command: mysqldump " + strings.Join(args, " "))

	return types_backup.DatabaseBackupInfo{
		DatabaseName:  formatBackupDisplayName(cfg),
		OutputFile:    cfg.OutputPath,
		FileSize:      0,
		FileSizeHuman: "0 B (dry-run)",
		Duration:      timer.Elapsed().String(),
		Status:        consts.BackupStatusDryRun,
		Warnings:      "Backup tidak dijalankan - mode dry-run aktif",
		StartTime:     startTime,
		EndTime:       time.Now(),
		ManifestFile:  "",
	}
}

// buildRealBackupInfo membuat DatabaseBackupInfo untuk backup yang sudah selesai.
func (e *Engine) buildRealBackupInfo(
	cfg types_backup.BackupExecutionConfig,
	writeResult *types_backup.BackupWriteResult,
	timer *timex.Timer,
	startTime time.Time,
	dbVersion string,
) types_backup.DatabaseBackupInfo {
	status := determineBackupStatus(writeResult, cfg, e.Log)

	duration := timer.Elapsed()
	endTime := time.Now()
	meta := e.generateBackupMetadata(cfg, writeResult, duration, startTime, endTime, status, dbVersion)

	manifestPath := ""
	if e.Config.Backup.Output.SaveBackupInfo {
		manifestPath = metadata.TrySaveBackupMetadata(meta, e.Log)
	}

	return (&metadata.DatabaseBackupInfoBuilder{
		DatabaseName: formatBackupDisplayName(cfg),
		OutputFile:   cfg.OutputPath,
		FileSize:     writeResult.FileSize,
		Duration:     duration,
		Status:       status,
		Warnings:     writeResult.StderrOutput,
		StartTime:    startTime,
		EndTime:      endTime,
		ManifestFile: manifestPath,
	}).Build()
}

// generateBackupMetadata membuat BackupMetadata object untuk sebuah backup.
func (e *Engine) generateBackupMetadata(
	cfg types_backup.BackupExecutionConfig,
	writeResult *types_backup.BackupWriteResult,
	duration time.Duration,
	startTime, endTime time.Time,
	status, dbVersion string,
) *types_backup.BackupMetadata {
	dbNames := []string{cfg.DBName}
	if cfg.IsMultiDB {
		dbNames = cfg.DBList
	}

	gtidStr := formatGTIDString(e.GTIDInfo)

	userGrantsPath := determineUserGrantsPath(e.Options.ExcludeUser, cfg.OutputPath)

	excludedDBs := getExcludedDatabases(cfg.BackupType, e.ExcludedDatabases)

	return metadata.GenerateBackupMetadata(types_backup.MetadataConfig{
		BackupFile:          cfg.OutputPath,
		BackupType:          cfg.BackupType,
		DatabaseNames:       dbNames,
		ExcludedDatabases:   excludedDBs,
		Hostname:            e.Options.Profile.DBInfo.HostName,
		FileSize:            writeResult.FileSize,
		Compressed:          e.Options.Compression.Enabled,
		CompressionType:     e.Options.Compression.Type,
		Encrypted:           e.Options.Encryption.Enabled,
		ExcludeData:         e.Options.Filter.ExcludeData,
		GTIDInfo:            gtidStr,
		BackupStatus:        status,
		StderrOutput:        writeResult.StderrOutput,
		Duration:            duration,
		StartTime:           startTime,
		EndTime:             endTime,
		Logger:              e.Log,
		ReplicationUser:     e.Config.Backup.Replication.ReplicationUser,
		ReplicationPassword: e.Config.Backup.Replication.ReplicationPassword,
		SourceHost:          e.Options.Profile.DBInfo.Host,
		SourcePort:          e.Options.Profile.DBInfo.Port,
		UserGrantsFile:      userGrantsPath,
		MysqldumpVersion:    ExtractMysqldumpVersion(writeResult.StderrOutput),
		MariaDBVersion:      dbVersion,
		Ticket:              e.Options.Ticket,
	})
}
