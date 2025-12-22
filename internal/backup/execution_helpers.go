package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/backup/helpers"
	"sfDBTools/internal/cleanup"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	pkghelper "sfDBTools/pkg/helper"
)

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

	// Build mysqldump args inline
	conn := helpers.DatabaseConn{
		Host:     s.BackupDBOptions.Profile.DBInfo.Host,
		Port:     s.BackupDBOptions.Profile.DBInfo.Port,
		User:     s.BackupDBOptions.Profile.DBInfo.User,
		Password: s.BackupDBOptions.Profile.DBInfo.Password,
	}
	filterOpts := helpers.FilterOptions{
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
	mysqldumpArgs := helpers.BuildMysqldumpArgs(s.Config.Backup.MysqlDumpArgs, conn, filterOpts, dbList, cfg.DBName, cfg.TotalDBFound)

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
	} else {
		s.Log.Error(fmt.Sprintf("gagal backup database %s: %v", cfg.DBName, err))
	}
}
