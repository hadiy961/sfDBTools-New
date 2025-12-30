package execution

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/backup/gtid"
	"sfDBTools/internal/backup/metadata"
	"sfDBTools/internal/backup/writer"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/errorlog"
	"sfDBTools/pkg/fsops"
	pkghelper "sfDBTools/pkg/helper"
)

type StateTracker interface {
	SetCurrentBackupFile(filePath string)
	ClearCurrentBackupFile()
}

type UserGrantsHooks interface {
	ExportUserGrantsIfNeeded(ctx context.Context, outputPath string, dbNames []string) string
	UpdateMetadataUserGrantsPath(outputPath string, userGrantsPath string)
}

type Engine struct {
	Log      applog.Logger
	Config   *appconfig.Config
	Options  *types_backup.BackupDBOptions
	ErrorLog *errorlog.ErrorLogger

	Client            *database.Client
	GTIDInfo          *gtid.GTIDInfo
	ExcludedDatabases []string

	State      StateTracker
	UserGrants UserGrantsHooks
}

func New(log applog.Logger, cfg *appconfig.Config, opts *types_backup.BackupDBOptions, errLog *errorlog.ErrorLogger) *Engine {
	return &Engine{Log: log, Config: cfg, Options: opts, ErrorLog: errLog}
}

// ExecuteAndBuildBackup runs a backup and produces DatabaseBackupInfo (incl. metadata/manifest).
func (e *Engine) ExecuteAndBuildBackup(ctx context.Context, cfg types_backup.BackupExecutionConfig) (types_backup.DatabaseBackupInfo, error) {
	timer := pkghelper.NewTimer()
	startTime := timer.StartTime()

	if e.State != nil {
		e.State.SetCurrentBackupFile(cfg.OutputPath)
		defer e.State.ClearCurrentBackupFile()
	}

	// Get DB version BEFORE backup (fresh connection)
	dbVersion := ""
	if e.Client != nil {
		if v, err := e.Client.GetVersion(ctx); err == nil {
			dbVersion = v
			e.Log.Debugf("Database version: %s", dbVersion)
		} else {
			e.Log.Warnf("Gagal mendapatkan database version: %v", err)
		}
	}

	// Build mysqldump args
	// Validate critical pointers to prevent nil dereference
	if e.Options == nil {
		return types_backup.DatabaseBackupInfo{}, fmt.Errorf("backup options tidak tersedia")
	}

	conn := DatabaseConn{
		Host:     e.Options.Profile.DBInfo.Host,
		Port:     e.Options.Profile.DBInfo.Port,
		User:     e.Options.Profile.DBInfo.User,
		Password: e.Options.Profile.DBInfo.Password,
	}
	filterOpts := FilterOptions{
		ExcludeData:      e.Options.Filter.ExcludeData,
		ExcludeDatabases: e.Options.Filter.ExcludeDatabases,
		IncludeDatabases: e.Options.Filter.IncludeDatabases,
		ExcludeSystem:    e.Options.Filter.ExcludeSystem,
		ExcludeDBFile:    e.Options.Filter.ExcludeDBFile,
		IncludeFile:      e.Options.Filter.IncludeFile,
	}
	var dbList []string
	if cfg.IsMultiDB {
		dbList = cfg.DBList
	}
	mysqldumpArgs := BuildMysqldumpArgs(e.Config.Backup.MysqlDumpArgs, conn, filterOpts, dbList, cfg.DBName, cfg.TotalDBFound)

	// DRY-RUN MODE
	if e.Options.DryRun {
		return e.buildDryRunBackupInfo(cfg, mysqldumpArgs, timer, startTime), nil
	}

	writeEngine := writer.New(e.Log, e.ErrorLog, e.Options)
	writeResult, err := writeEngine.ExecuteMysqldumpWithPipe(ctx, mysqldumpArgs, cfg.OutputPath, e.Options.Compression.Enabled, e.Options.Compression.Type)
	if err != nil {
		// If canceled, don't treat as regular failure
		if ctx.Err() != nil {
			e.Log.Warnf("Backup database %s dibatalkan", cfg.DBName)
			cleanupFailedBackup(cfg.OutputPath, e.Log)
			return types_backup.DatabaseBackupInfo{}, err
		}

		// Fast retry for common TLS/SSL mismatch (client requires SSL but server doesn't support it).
		// This often happens due to default client config enforcing SSL.
		if writeResult != nil {
			if IsSSLMismatchRequiredButServerNoSupport(writeResult.StderrOutput) {
				if newArgs, added := AddDisableSSLArgs(mysqldumpArgs); added {
					e.Log.Warn("mysqldump gagal karena SSL required tapi server tidak support SSL. Retry dengan --skip-ssl...")
					cleanupFailedBackup(cfg.OutputPath, e.Log)
					writeResult2, err2 := writeEngine.ExecuteMysqldumpWithPipe(ctx, newArgs, cfg.OutputPath, e.Options.Compression.Enabled, e.Options.Compression.Type)
					if err2 == nil {
						return e.buildBackupInfoFromResult(ctx, cfg, writeResult2, timer, startTime, dbVersion), nil
					}
					writeResult = writeResult2
					err = err2
				}
			}
		}

		// Fast retry for common "exit status 2" cases (unknown option/variable).
		// This usually happens when config contains mysqldump flags unsupported by the installed client.
		if writeResult != nil {
			if newArgs, removed, ok := RemoveUnsupportedMysqldumpOption(mysqldumpArgs, writeResult.StderrOutput); ok {
				e.Log.Warnf("mysqldump gagal karena opsi tidak didukung (%s). Retry tanpa opsi tersebut...", removed)
				cleanupFailedBackup(cfg.OutputPath, e.Log)
				writeResult2, err2 := writeEngine.ExecuteMysqldumpWithPipe(ctx, newArgs, cfg.OutputPath, e.Options.Compression.Enabled, e.Options.Compression.Type)
				if err2 == nil {
					return e.buildBackupInfoFromResult(ctx, cfg, writeResult2, timer, startTime, dbVersion), nil
				}
				// If retry fails, return the retry error (still logged by handler below).
				writeResult = writeResult2
				err = err2
			}
		}
		e.handleBackupError(err, cfg, writeResult)
		return types_backup.DatabaseBackupInfo{}, err
	}

	return e.buildBackupInfoFromResult(ctx, cfg, writeResult, timer, startTime, dbVersion), nil
}

func (e *Engine) handleBackupError(err error, cfg types_backup.BackupExecutionConfig, writeResult *types_backup.BackupWriteResult) {
	stderrDetail := ""
	if writeResult != nil {
		stderrDetail = writeResult.StderrOutput
	}

	if e.ErrorLog != nil {
		logMeta := map[string]interface{}{"type": cfg.BackupType + "_backup", "file": cfg.OutputPath}
		if !cfg.IsMultiDB {
			logMeta["database"] = cfg.DBName
		}
		e.ErrorLog.LogWithOutput(logMeta, stderrDetail, err)
	}

	cleanupFailedBackup(cfg.OutputPath, e.Log)

	if cfg.IsMultiDB {
		e.Log.Errorf("Gagal menjalankan mysqldump: %v", err)
	} else {
		e.Log.Error(fmt.Sprintf("gagal backup database %s: %v", cfg.DBName, err))
	}
}

func cleanupFailedBackup(filePath string, logger applog.Logger) {
	if fsops.FileExists(filePath) {
		logger.Infof("Menghapus file backup yang gagal: %s", filePath)
		if err := fsops.RemoveFile(filePath); err != nil {
			logger.Warnf("Gagal menghapus file backup yang gagal: %v", err)
		}
	}
}

// ExecuteBackupLoop runs backup across multiple databases.
func (e *Engine) ExecuteBackupLoop(ctx context.Context, databases []string, config types_backup.BackupLoopConfig, outputPathFunc func(dbName string) (string, error)) types_backup.BackupLoopResult {
	result := types_backup.BackupLoopResult{
		BackupInfos: make([]types_backup.DatabaseBackupInfo, 0),
		FailedDBs:   make([]types_backup.FailedDatabaseInfo, 0),
		Errors:      make([]string, 0),
	}

	if len(databases) == 0 {
		e.Log.Warn("Tidak ada database yang dipilih untuk backup")
		result.Errors = append(result.Errors, "tidak ada database yang dipilih")
		return result
	}

	for idx, dbName := range databases {
		if ctx.Err() != nil {
			e.Log.Warn("Proses backup dibatalkan")
			result.Errors = append(result.Errors, "Backup dibatalkan oleh user")
			break
		}

		e.Log.Infof("[%d/%d] Backup database: %s", idx+1, len(databases), dbName)

		outputPath, err := outputPathFunc(dbName)
		if err != nil {
			msg := fmt.Sprintf("gagal generate path untuk %s: %v", dbName, err)
			e.Log.Error(msg)
			result.FailedDBs = append(result.FailedDBs, types_backup.FailedDatabaseInfo{DatabaseName: dbName, Error: msg})
			result.Failed++
			continue
		}

		backupInfo, err := e.ExecuteAndBuildBackup(ctx, types_backup.BackupExecutionConfig{
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
		if e.UserGrants != nil {
			if config.Mode == consts.ModeSeparated || config.Mode == consts.ModeSeparate || config.Mode == consts.ModeSingle {
				path := e.UserGrants.ExportUserGrantsIfNeeded(ctx, outputPath, []string{dbName})
				if e.Config.Backup.Output.SaveBackupInfo {
					e.UserGrants.UpdateMetadataUserGrantsPath(outputPath, path)
				}
			}
		}
	}

	return result
}

func (e *Engine) buildDryRunBackupInfo(cfg types_backup.BackupExecutionConfig, mysqldumpArgs []string, timer *pkghelper.Timer, startTime time.Time) types_backup.DatabaseBackupInfo {
	backupDuration := timer.Elapsed()
	endTime := time.Now()

	if cfg.IsMultiDB {
		e.Log.Info("[DRY-RUN] Akan backup database: " + strings.Join(cfg.DBList, ", "))
	} else {
		e.Log.Infof("[DRY-RUN] Akan backup database: %s", cfg.DBName)
	}
	e.Log.Info("[DRY-RUN] Output file: " + cfg.OutputPath)
	e.Log.Debug("[DRY-RUN] Mysqldump command: mysqldump " + strings.Join(mysqldumpArgs, " "))
	e.Log.Info("[DRY-RUN] Backup tidak dijalankan (dry-run mode aktif)")

	displayName := formatBackupDisplayName(cfg)

	return types_backup.DatabaseBackupInfo{
		DatabaseName:  displayName,
		OutputFile:    cfg.OutputPath,
		FileSize:      0,
		FileSizeHuman: "0 B (dry-run)",
		Duration:      backupDuration.String(),
		Status:        consts.BackupStatusDryRun,
		Warnings:      "Backup tidak dijalankan - mode dry-run aktif",
		StartTime:     startTime,
		EndTime:       endTime,
		ManifestFile:  "",
	}
}

func (e *Engine) buildBackupInfoFromResult(ctx context.Context, cfg types_backup.BackupExecutionConfig, writeResult *types_backup.BackupWriteResult, timer *pkghelper.Timer, startTime time.Time, dbVersion string) types_backup.DatabaseBackupInfo {
	fileSize := writeResult.FileSize
	stderrOutput := writeResult.StderrOutput
	backupDuration := timer.Elapsed()
	endTime := time.Now()

	status := consts.BackupStatusSuccess
	if stderrOutput != "" {
		status = consts.BackupStatusSuccessWithWarnings
		if !cfg.IsMultiDB {
			e.Log.Warningf("Database %s backup dengan warning: %s", cfg.DBName, stderrOutput)
		}
	} else if !cfg.IsMultiDB {
		e.Log.Infof("✓ Database %s berhasil di-backup", cfg.DBName)
	}

	meta := e.generateBackupMetadata(ctx, cfg, writeResult, backupDuration, startTime, endTime, status, dbVersion)

	manifestPath := ""
	if e.Config.Backup.Output.SaveBackupInfo {
		manifestPath = metadata.TrySaveBackupMetadata(meta, e.Log)
	}

	displayName := formatBackupDisplayName(cfg)

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

func formatBackupDisplayName(cfg types_backup.BackupExecutionConfig) string {
	if cfg.IsMultiDB {
		return fmt.Sprintf("Combined backup (%d databases)", len(cfg.DBList))
	}
	return cfg.DBName
}

func (e *Engine) generateBackupMetadata(ctx context.Context, cfg types_backup.BackupExecutionConfig, writeResult *types_backup.BackupWriteResult, duration time.Duration, startTime, endTime time.Time, status string, dbVersion string) *types_backup.BackupMetadata {
	var dbNames []string
	if cfg.IsMultiDB {
		dbNames = cfg.DBList
	} else {
		dbNames = []string{cfg.DBName}
	}

	gtidStr := ""
	if e.GTIDInfo != nil {
		if e.GTIDInfo.GTIDBinlog != "" {
			gtidStr = e.GTIDInfo.GTIDBinlog
		} else {
			gtidStr = fmt.Sprintf("File=%s, Pos=%d", e.GTIDInfo.MasterLogFile, e.GTIDInfo.MasterLogPos)
		}
	}

	userGrantsPath := ""
	if !e.Options.ExcludeUser {
		userGrantsPath = metadata.GenerateUserFilePath(cfg.OutputPath)
	}

	excludedDBs := []string{}
	if cfg.BackupType == consts.ModeAll {
		excludedDBs = e.ExcludedDatabases
	}

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
