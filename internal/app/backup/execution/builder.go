// File : internal/backup/execution/builder.go
// Deskripsi : Builder functions untuk DatabaseBackupInfo dan metadata generation
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-31
// Last Modified : 20 Januari 2026

package execution

import (
	"strings"
	"time"

	"sfdbtools/internal/app/backup/metadata"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/timex"
)

// maskDumpArgsForLog menghindari kebocoran kredensial saat menampilkan dump args ke log.
// Saat ini yang di-mask: --password=... dan form dua-arg "--password" "..."
func maskDumpArgsForLog(args []string) []string {
	if len(args) == 0 {
		return args
	}
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		al := strings.ToLower(strings.TrimSpace(a))
		if strings.HasPrefix(al, "--password=") {
			out = append(out, "--password=***")
			continue
		}
		if al == "--password" {
			out = append(out, "--password")
			// mask next token value if present
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				out = append(out, "***")
				i++
			}
			continue
		}
		out = append(out, a)
	}
	return out
}

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
	e.Log.Debug("[DRY-RUN] Dump tool (auto): mariadb-dump (fallback mysqldump) " + strings.Join(maskDumpArgsForLog(args), " "))

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
		manifestPath = metadata.TrySaveBackupMetadata(meta, e.Config.Backup.Output.MetadataPermissions, e.Log)
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
		// SECURITY: Jangan pernah menyimpan password replikasi ke metadata/manifest.
		ReplicationPassword: "",
		SourceHost:          e.Options.Profile.DBInfo.Host,
		SourcePort:          e.Options.Profile.DBInfo.Port,
		UserGrantsFile:      userGrantsPath,
		MysqldumpVersion:    ExtractMysqldumpVersion(writeResult.StderrOutput),
		MariaDBVersion:      dbVersion,
		Ticket:              e.Options.Ticket,
	})
}
