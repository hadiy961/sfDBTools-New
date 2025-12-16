// File : internal/backup/service_helpers.go
// Deskripsi : Service helper methods untuk backup operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package backup

import (
	"context"
	"fmt"
	"path/filepath"
	"sfDBTools/internal/backup/filehelper"
	"sfDBTools/internal/backup/metadata"
	"sfDBTools/internal/cleanup"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/backuphelper"
	"sfDBTools/pkg/compress"
	pkghelper "sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"time"
)

// =============================================================================
// Path Generation Helpers
// =============================================================================

// generateFullBackupPath membuat full path untuk backup file
func (s *Service) generateFullBackupPath(dbName string, mode string) (string, error) {
	// Build compression settings inline
	compressionType := s.BackupDBOptions.Compression.Type
	if !s.BackupDBOptions.Compression.Enabled {
		compressionType = ""
	}
	compressionSettings := types_backup.CompressionSettings{
		Type:    compress.CompressionType(compressionType),
		Enabled: s.BackupDBOptions.Compression.Enabled,
		Level:   s.BackupDBOptions.Compression.Level,
	}

	// Untuk mode separated, gunakan IP address instead of hostname
	hostIdentifier := s.BackupDBOptions.Profile.DBInfo.HostName
	if mode == "separated" || mode == "separate" {
		hostIdentifier = s.BackupDBOptions.Profile.DBInfo.Host
	}

	filename, err := pkghelper.GenerateBackupFilename(
		dbName,
		mode,
		hostIdentifier,
		compressionSettings.Type,
		s.BackupDBOptions.Encryption.Enabled,
	)
	if err != nil {
		return "", err
	}

	return filepath.Join(s.BackupDBOptions.OutputDir, filename), nil
}

// =============================================================================
// Backup Loop Helpers
// =============================================================================

// executeBackupLoop menjalankan backup untuk multiple databases dengan pattern yang sama
func (s *Service) executeBackupLoop(ctx context.Context, databases []string, config types_backup.BackupLoopConfig, outputPathFunc func(dbName string) (string, error)) types_backup.BackupLoopResult {
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

		s.Log.Debugf("Backup file: %s", outputPath)

		// Execute backup
		backupInfo, err := s.executeAndBuildBackup(ctx, types_backup.BackupExecutionConfig{
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
			path := s.exportUserGrantsIfNeeded(ctx, outputPath, []string{dbName})
			if s.Config.Backup.Output.SaveBackupInfo {
				s.updateMetadataUserGrantsPath(outputPath, path)
			}
		}
	}

	return result
}

// =============================================================================
// Backup Execution Helpers
// =============================================================================

// executeAndBuildBackup menjalankan backup dan generate metadata + backup info
func (s *Service) executeAndBuildBackup(ctx context.Context, cfg types_backup.BackupExecutionConfig) (types.DatabaseBackupInfo, error) {
	timer := pkghelper.NewTimer()
	startTime := timer.StartTime()

	s.SetCurrentBackupFile(cfg.OutputPath)
	defer s.ClearCurrentBackupFile()

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

	// Execute backup
	writeResult, err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, cfg.OutputPath, s.BackupDBOptions.Compression.Enabled, s.BackupDBOptions.Compression.Type)
	if err != nil {
		s.handleBackupError(err, cfg, writeResult)
		return types.DatabaseBackupInfo{}, err
	}

	return s.buildBackupInfoFromResult(ctx, cfg, writeResult, timer, startTime), nil
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

// buildBackupInfoFromResult membangun DatabaseBackupInfo dari result
func (s *Service) buildBackupInfoFromResult(ctx context.Context, cfg types_backup.BackupExecutionConfig, writeResult *types_backup.BackupWriteResult, timer *pkghelper.Timer, startTime time.Time) types.DatabaseBackupInfo {
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
	meta := s.generateBackupMetadata(ctx, cfg, writeResult, backupDuration, startTime, endTime, status)

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
func (s *Service) generateBackupMetadata(ctx context.Context, cfg types_backup.BackupExecutionConfig, writeResult *types_backup.BackupWriteResult, duration time.Duration, startTime, endTime time.Time, status string) *types_backup.BackupMetadata {
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

	// Get DB version
	dbVersion := ""
	if s.Client != nil {
		if v, err := s.Client.GetVersion(ctx); err == nil {
			dbVersion = v
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
	})
}

// =============================================================================
// GTID and User Grants Helpers
// =============================================================================

// captureAndSaveGTID mengambil dan menyimpan GTID info jika diperlukan
func (s *Service) captureAndSaveGTID(ctx context.Context, backupFilePath string) error {
	if !s.BackupDBOptions.CaptureGTID {
		return nil
	}

	s.Log.Info("Mengambil informasi GTID sebelum backup...")
	gtidInfo, err := s.Client.GetFullGTIDInfo(ctx)
	if err != nil {
		s.Log.Warnf("Gagal mendapatkan GTID: %v", err)
		return nil
	}

	s.Log.Infof("GTID berhasil diambil: File=%s, Pos=%d", gtidInfo.MasterLogFile, gtidInfo.MasterLogPos)

	// Simpan GTID info ke service untuk dimasukkan ke metadata nanti
	s.gtidInfo = gtidInfo

	return nil
}

// getTotalDatabaseCount mengambil total database dari server
func (s *Service) getTotalDatabaseCount(ctx context.Context, dbFiltered []string) int {
	allDatabases, err := s.Client.GetDatabaseList(ctx)
	totalDBFound := len(allDatabases)
	if err != nil {
		s.Log.Warnf("Gagal mendapatkan total database: %v, menggunakan fallback", err)
		totalDBFound = len(dbFiltered)
	}
	return totalDBFound
}

// exportUserGrantsIfNeeded export user grants jika diperlukan
// Delegates to metadata.ExportUserGrantsIfNeededWithLogging dengan BackupDBOptions.ExcludeUser
func (s *Service) exportUserGrantsIfNeeded(ctx context.Context, referenceBackupFile string, databases []string) string {
	return metadata.ExportUserGrantsIfNeededWithLogging(ctx, s.Client, s.Log, referenceBackupFile, s.BackupDBOptions.ExcludeUser, databases)
}

// updateMetadataUserGrantsPath update metadata dengan actual user grants path
func (s *Service) updateMetadataUserGrantsPath(backupFilePath string, userGrantsPath string) {
	if err := metadata.UpdateMetadataUserGrantsFile(backupFilePath, userGrantsPath, s.Log); err != nil {
		s.Log.Warnf("Gagal update metadata user grants path: %v", err)
	}
}

// =============================================================================
// Database Selection Helpers
// =============================================================================

// selectDatabaseAndBuildList menangani database selection dan companion databases logic
func (s *Service) selectDatabaseAndBuildList(ctx context.Context, client interface {
	GetDatabaseList(context.Context) ([]string, error)
}, selectedDBName string, dbFiltered []string, mode string) ([]string, string, map[string]bool, error) {

	allDatabases, listErr := client.GetDatabaseList(ctx)
	if listErr != nil {
		return nil, "", nil, fmt.Errorf("gagal mengambil daftar database: %w", listErr)
	}

	selectedDB := selectedDBName
	if selectedDB == "" {
		candidates := backuphelper.FilterCandidatesByMode(dbFiltered, mode)

		if len(candidates) == 0 {
			return nil, "", nil, fmt.Errorf("tidak ada database yang tersedia untuk dipilih")
		}

		choice, choiceErr := input.ShowMenu("Pilih database yang akan di-backup:", candidates)
		if choiceErr != nil {
			return nil, "", nil, choiceErr
		}
		selectedDB = candidates[choice-1]
	}

	if !pkghelper.StringSliceContainsFold(allDatabases, selectedDB) {
		return nil, "", nil, fmt.Errorf("database %s tidak ditemukan di server", selectedDB)
	}

	companionDbs := []string{selectedDB}
	companionStatus := map[string]bool{selectedDB: true}

	// Add companion databases - consolidated loop for all suffixes
	for suffix, enabled := range map[string]bool{
		"_dmart":   s.BackupDBOptions.IncludeDmart,
		"_temp":    s.BackupDBOptions.IncludeTemp,
		"_archive": s.BackupDBOptions.IncludeArchive,
	} {
		if !enabled {
			continue
		}

		dbName := selectedDB + suffix
		exists := pkghelper.StringSliceContainsFold(allDatabases, dbName)

		if exists {
			s.Log.Infof("Menambahkan database companion: %s", dbName)
			companionDbs = append(companionDbs, dbName)
		} else {
			s.Log.Warnf("Database %s tidak ditemukan, melewati", dbName)
		}
		companionStatus[dbName] = exists
	}

	return companionDbs, selectedDB, companionStatus, nil
}
