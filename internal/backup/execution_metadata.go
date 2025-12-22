package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/backup/metadata"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/backuphelper"
	"sfDBTools/pkg/consts"
	"time"
)

// generateBackupMetadata generates metadata config untuk backup
func (s *Service) generateBackupMetadata(ctx context.Context, cfg types_backup.BackupExecutionConfig, writeResult *types_backup.BackupWriteResult, duration time.Duration, startTime, endTime time.Time, status string, dbVersion string) *types_backup.BackupMetadata {
	var dbNames []string
	if cfg.IsMultiDB {
		dbNames = cfg.DBList
	} else {
		dbNames = []string{cfg.DBName}
	}

	// Format GTID info
	gtidStr := ""
	if s.gtidInfo != nil {
		if s.gtidInfo.GTIDBinlog != "" {
			gtidStr = s.gtidInfo.GTIDBinlog
		} else {
			gtidStr = fmt.Sprintf("File=%s, Pos=%d", s.gtidInfo.MasterLogFile, s.gtidInfo.MasterLogPos)
		}
	}

	// User grants path
	userGrantsPath := ""
	if !s.BackupDBOptions.ExcludeUser {
		userGrantsPath = metadata.GenerateUserFilePath(cfg.OutputPath)
	}

	// Excluded databases
	excludedDBs := []string{}
	if cfg.BackupType == consts.ModeAll && s.excludedDatabases != nil {
		excludedDBs = s.excludedDatabases
	}

	return metadata.GenerateBackupMetadata(types_backup.MetadataConfig{
		BackupFile:          cfg.OutputPath,
		BackupType:          cfg.BackupType,
		DatabaseNames:       dbNames,
		ExcludedDatabases:   excludedDBs,
		Hostname:            s.BackupDBOptions.Profile.DBInfo.HostName,
		FileSize:            writeResult.FileSize,
		Compressed:          s.BackupDBOptions.Compression.Enabled,
		CompressionType:     s.BackupDBOptions.Compression.Type,
		Encrypted:           s.BackupDBOptions.Encryption.Enabled,
		ExcludeData:         s.BackupDBOptions.Filter.ExcludeData,
		GTIDInfo:            gtidStr,
		BackupStatus:        status,
		StderrOutput:        writeResult.StderrOutput,
		Duration:            duration,
		StartTime:           startTime,
		EndTime:             endTime,
		Logger:              s.Log,
		ReplicationUser:     s.Config.Backup.Replication.ReplicationUser,
		ReplicationPassword: s.Config.Backup.Replication.ReplicationPassword,
		SourceHost:          s.BackupDBOptions.Profile.DBInfo.Host,
		SourcePort:          s.BackupDBOptions.Profile.DBInfo.Port,
		UserGrantsFile:      userGrantsPath,
		MysqldumpVersion:    backuphelper.ExtractMysqldumpVersion(writeResult.StderrOutput),
		MariaDBVersion:      dbVersion,
		Ticket:              s.BackupDBOptions.Ticket,
	})
}
