package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/backup/filehelper"
	"sfDBTools/internal/backup/metadata"
	"sfDBTools/internal/cleanup"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/backuphelper"
	pkghelper "sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"
	"time"
)

// =============================================================================
// Backup Loop Helpers
// =============================================================================

// ExecuteBackupLoop menjalankan backup untuk multiple databases dengan pattern yang sama
func (s *Service) ExecuteBackupLoop(ctx context.Context, databases []string, config types_backup.BackupLoopConfig, outputPathFunc func(dbName string) (string, error)) types_backup.BackupLoopResult {
	result := types_backup.BackupLoopResult{
		BackupInfos: make([]types.DatabaseBackupInfo, 0),
		FailedDBs:   make([]types_backup.FailedDatabaseInfo, 0),
		Errors:      make([]string, 0),
	}

	if len(databases) == 0 {
		s.Log.Warn("Tidak ada database yang dipilih untuk backup")
		result.Errors = append(result.Errors, "tidak ada database yang dipilih")
		return result
	}

	for idx, dbName := range databases {
		if ctx.Err() != nil {
			s.Log.Warn("Proses backup dibatalkan")
			result.Errors = append(result.Errors, "Backup dibatalkan oleh user")
			break
		}

		s.Log.Infof("[%d/%d] Backup database: %s", idx+1, len(databases), dbName)

		// Generate output path
		outputPath, err := outputPathFunc(dbName)
		if err != nil {
			msg := fmt.Sprintf("gagal generate path untuk %s: %v", dbName, err)
			s.Log.Error(msg)
			result.FailedDBs = append(result.FailedDBs, types_backup.FailedDatabaseInfo{DatabaseName: dbName, Error: msg})
			result.Failed++
			continue
		}

		// Execute backup
		backupInfo, err := s.ExecuteAndBuildBackup(ctx, types_backup.BackupExecutionConfig{
			DBName:       dbName,
			OutputPath:   outputPath,
			BackupType:   config.BackupType,
			TotalDBFound: config.TotalDBs,
			IsMultiDB:    false,
		})

		if err != nil {
			result.FailedDBs = append(result.FailedDBs, types_backup.FailedDatabaseInfo{DatabaseName: dbName, Error: err.Error()})
			result.Failed++
			continue
		}

		result.BackupInfos = append(result.BackupInfos, backupInfo)
		result.Success++

		// Export user grants for separated/single modes
		if config.Mode == "separated" || config.Mode == "separate" || config.Mode == "single" {
			path := s.ExportUserGrantsIfNeeded(ctx, outputPath, []string{dbName})
			if s.Config.Backup.Output.SaveBackupInfo {
				s.UpdateMetadataUserGrantsPath(outputPath, path)
			}
		}
	}

	return result
}

// =============================================================================
// Backup Execution Helpers
// =============================================================================

// ExecuteAndBuildBackup menjalankan backup dan generate metadata + backup info
func (s *Service) ExecuteAndBuildBackup(ctx context.Context, cfg types_backup.BackupExecutionConfig) (types.DatabaseBackupInfo, error) {
	timer := pkghelper.NewTimer()
	startTime := timer.StartTime()

	s.SetCurrentBackupFile(cfg.OutputPath)
	defer s.ClearCurrentBackupFile()

	// Get DB version SEBELUM backup (koneksi masih fresh)
	dbVersion := ""
	if s.Client != nil {
		if v, err := s.Client.GetVersion(ctx); err == nil {
			dbVersion = v
			s.Log.Debugf("Database version: %s", dbVersion)
		} else {
			s.Log.Warnf("Gagal mendapatkan database version: %v", err)
		}
	}

	// Debug: Log ExcludeData value
	// s.Log.Debugf("ExcludeData flag value: %v", s.BackupDBOptions.Filter.ExcludeData)

	// Build mysqldump args inline
	conn := backuphelper.DatabaseConn{
		Host:     s.BackupDBOptions.Profile.DBInfo.Host,
		Port:     s.BackupDBOptions.Profile.DBInfo.Port,
		User:     s.BackupDBOptions.Profile.DBInfo.User,
		Password: s.BackupDBOptions.Profile.DBInfo.Password,
	}
	filterOpts := backuphelper.FilterOptions{
		ExcludeData:      s.BackupDBOptions.Filter.ExcludeData,
		ExcludeDatabases: s.BackupDBOptions.Filter.ExcludeDatabases,
		IncludeDatabases: s.BackupDBOptions.Filter.IncludeDatabases,
		ExcludeSystem:    s.BackupDBOptions.Filter.ExcludeSystem,
		ExcludeDBFile:    s.BackupDBOptions.Filter.ExcludeDBFile,
		IncludeFile:      s.BackupDBOptions.Filter.IncludeFile,
	}
	var dbList []string
	if cfg.IsMultiDB {
		dbList = cfg.DBList
	}
	mysqldumpArgs := backuphelper.BuildMysqldumpArgs(s.Config.Backup.MysqlDumpArgs, conn, filterOpts, dbList, cfg.DBName, cfg.TotalDBFound)

	// DRY-RUN MODE: Skip actual backup execution
	if s.BackupDBOptions.DryRun {
		return s.buildDryRunBackupInfo(cfg, mysqldumpArgs, timer, startTime), nil
	}

	// Execute backup
	writeResult, err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, cfg.OutputPath, s.BackupDBOptions.Compression.Enabled, s.BackupDBOptions.Compression.Type)
	if err != nil {
		// Jika error karena context canceled (shutdown), jangan log sebagai error
		if ctx.Err() != nil {
			s.Log.Warnf("Backup database %s dibatalkan", cfg.DBName)
			cleanup.CleanupFailedBackup(cfg.OutputPath, s.Log)
			return types.DatabaseBackupInfo{}, err
		}
		s.handleBackupError(err, cfg, writeResult)
		return types.DatabaseBackupInfo{}, err
	}

	return s.buildBackupInfoFromResult(ctx, cfg, writeResult, timer, startTime, dbVersion), nil
}

// handleBackupError menangani error dari backup execution
func (s *Service) handleBackupError(err error, cfg types_backup.BackupExecutionConfig, writeResult *types_backup.BackupWriteResult) {
	stderrDetail := ""
	if writeResult != nil {
		stderrDetail = writeResult.StderrOutput
	}

	if s.ErrorLog != nil {
		logMeta := map[string]interface{}{"type": cfg.BackupType + "_backup", "file": cfg.OutputPath}
		if !cfg.IsMultiDB {
			logMeta["database"] = cfg.DBName
		}
		s.ErrorLog.LogWithOutput(logMeta, stderrDetail, err)
	}

	cleanup.CleanupFailedBackup(cfg.OutputPath, s.Log)

	if cfg.IsMultiDB {
		s.Log.Errorf("Gagal menjalankan mysqldump: %v", err)
		if stderrDetail != "" {
			ui.PrintHeader("ERROR : Mysqldump Gagal dijalankan")
			ui.PrintSubHeader("Detail Error dari mysqldump:")
			ui.PrintError(stderrDetail)
		}
	} else {
		ui.PrintError(fmt.Sprintf("✗ Database %s gagal di-backup", cfg.DBName))
		s.Log.Error(fmt.Sprintf("gagal backup database %s: %v", cfg.DBName, err))
	}
}

// buildDryRunBackupInfo membangun dummy DatabaseBackupInfo untuk dry-run mode
func (s *Service) buildDryRunBackupInfo(cfg types_backup.BackupExecutionConfig, mysqldumpArgs []string, timer *pkghelper.Timer, startTime time.Time) types.DatabaseBackupInfo {
	backupDuration := timer.Elapsed()
	endTime := time.Now()

	// Log dry-run info
	if cfg.IsMultiDB {
		s.Log.Info("[DRY-RUN] Akan backup database: " + strings.Join(cfg.DBList, ", "))
	} else {
		s.Log.Infof("[DRY-RUN] Akan backup database: %s", cfg.DBName)
	}
	s.Log.Info("[DRY-RUN] Output file: " + cfg.OutputPath)
	s.Log.Debug("[DRY-RUN] Mysqldump command: mysqldump " + strings.Join(mysqldumpArgs, " "))
	s.Log.Info("[DRY-RUN] Backup tidak dijalankan (dry-run mode aktif)")

	// Format display name
	displayName := cfg.DBName
	if cfg.IsMultiDB {
		displayName = fmt.Sprintf("Combined backup (%d databases)", len(cfg.DBList))
	}

	return types.DatabaseBackupInfo{
		DatabaseName:  displayName,
		OutputFile:    cfg.OutputPath,
		FileSize:      0,
		FileSizeHuman: "0 B (dry-run)",
		Duration:      backupDuration.String(),
		Status:        "dry-run",
		Warnings:      "Backup tidak dijalankan - mode dry-run aktif",
		StartTime:     startTime,
		EndTime:       endTime,
		ManifestFile:  "",
	}
}

// buildBackupInfoFromResult membangun DatabaseBackupInfo dari result
func (s *Service) buildBackupInfoFromResult(ctx context.Context, cfg types_backup.BackupExecutionConfig, writeResult *types_backup.BackupWriteResult, timer *pkghelper.Timer, startTime time.Time, dbVersion string) types.DatabaseBackupInfo {
	fileSize := writeResult.FileSize
	stderrOutput := writeResult.StderrOutput
	backupDuration := timer.Elapsed()
	endTime := time.Now()

	// Determine status
	status := "success"
	if stderrOutput != "" {
		status = "success_with_warnings"
		if !cfg.IsMultiDB {
			s.Log.Warningf("Database %s backup dengan warning: %s", cfg.DBName, stderrOutput)
		}
	} else if !cfg.IsMultiDB {
		s.Log.Infof("✓ Database %s berhasil di-backup", cfg.DBName)
	}

	// Get metadata
	meta := s.generateBackupMetadata(ctx, cfg, writeResult, backupDuration, startTime, endTime, status, dbVersion)

	// Save and get manifest path
	manifestPath := ""
	if s.Config.Backup.Output.SaveBackupInfo {
		manifestPath = metadata.TrySaveBackupMetadata(meta, s.Log)
	}

	// Format display name
	displayName := cfg.DBName
	if cfg.IsMultiDB {
		displayName = fmt.Sprintf("Combined backup (%d databases)", len(cfg.DBList))
	}

	return (&metadata.DatabaseBackupInfoBuilder{
		DatabaseName: displayName,
		OutputFile:   cfg.OutputPath,
		FileSize:     fileSize,
		Duration:     backupDuration,
		Status:       status,
		Warnings:     stderrOutput,
		StartTime:    startTime,
		EndTime:      endTime,
		ManifestFile: manifestPath,
	}).Build()
}

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
		userGrantsPath = filehelper.GenerateUserFilePath(cfg.OutputPath)
	}

	// Excluded databases
	excludedDBs := []string{}
	if cfg.BackupType == "all" && s.excludedDatabases != nil {
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
