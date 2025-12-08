// File : internal/backup/helpers.go
// Deskripsi : Helper functions untuk backup operations (konsolidasi dari berbagai helper files)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package backup

import (
	"context"
	"fmt"
	"path/filepath"
	"sfDBTools/internal/backup/helper"
	"sfDBTools/internal/backup/metadata"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/backuphelper"
	pkghelper "sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"strings"
	"time"
)

// =============================================================================
// Path Generation Helpers
// =============================================================================

// generateFullBackupPath membuat full path untuk backup file
func (s *Service) generateFullBackupPath(dbName string, mode string) (string, error) {
	compressionSettings := helper.NewCompressionSettings(
		s.BackupDBOptions.Compression.Enabled,
		s.BackupDBOptions.Compression.Type,
		s.BackupDBOptions.Compression.Level,
	)

	filename, err := pkghelper.GenerateBackupFilename(
		dbName,
		mode,
		s.BackupDBOptions.Profile.DBInfo.HostName,
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
func (s *Service) executeBackupLoop(
	ctx context.Context,
	databases []string,
	config types_backup.BackupLoopConfig,
	outputPathFunc func(dbName string) (string, error),
) types_backup.BackupLoopResult {
	result := types_backup.BackupLoopResult{
		BackupInfos: []types.DatabaseBackupInfo{},
		FailedDBs:   []types_backup.FailedDatabaseInfo{},
		Errors:      []string{},
	}

	totalDatabases := len(databases)
	if totalDatabases == 0 {
		s.Log.Warn("Tidak ada database yang dipilih untuk backup")
		result.Errors = append(result.Errors, "tidak ada database yang dipilih")
		return result
	}

	for idx, dbName := range databases {
		// Graceful shutdown check
		if err := ctx.Err(); err != nil {
			s.Log.Warn("Proses backup dibatalkan oleh user")
			result.Errors = append(result.Errors, "Backup dibatalkan oleh user")
			break
		}

		progress := fmt.Sprintf("[%d/%d]", idx+1, totalDatabases)
		s.Log.Infof("%s Backup database: %s", progress, dbName)

		// Generate output path
		outputPath, err := outputPathFunc(dbName)
		if err != nil {
			errorMsg := fmt.Sprintf("gagal generate path untuk %s: %v", dbName, err)
			s.Log.Error(errorMsg)
			result.FailedDBs = append(result.FailedDBs, types_backup.FailedDatabaseInfo{
				DatabaseName: dbName,
				Error:        errorMsg,
			})
			result.Failed++
			continue
		}

		s.Log.Debugf("Backup file: %s", outputPath)

		// Execute backup
		backupInfo, execErr := s.executeAndBuildBackup(ctx, types_backup.BackupExecutionConfig{
			DBName:       dbName,
			OutputPath:   outputPath,
			BackupType:   config.BackupType,
			TotalDBFound: config.TotalDBs,
			IsMultiDB:    false,
		})

		if execErr != nil {
			result.FailedDBs = append(result.FailedDBs, types_backup.FailedDatabaseInfo{
				DatabaseName: dbName,
				Error:        execErr.Error(),
			})
			result.Failed++
			continue
		}

		result.BackupInfos = append(result.BackupInfos, backupInfo)
		result.Success++
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

	// Set current backup file untuk graceful shutdown
	s.SetCurrentBackupFile(cfg.OutputPath)
	defer s.ClearCurrentBackupFile()

	// Build mysqldump args
	mysqldumpArgs := s.buildMysqldumpArgsFromConfig(cfg, s.BackupDBOptions.Profile, s.BackupDBOptions.Filter, s.Config.Backup.MysqlDumpArgs)

	// Execute backup
	writeResult, err := s.executeMysqldumpWithPipe(
		ctx,
		mysqldumpArgs,
		cfg.OutputPath,
		s.BackupDBOptions.Compression.Enabled,
		s.BackupDBOptions.Compression.Type,
	)

	// Handle error
	if err != nil {
		errorHandler := NewBackupErrorHandler(s.Log, s.ErrorLog, true)
		stderrDetail := ""
		if writeResult != nil && writeResult.StderrOutput != "" {
			stderrDetail = writeResult.StderrOutput
		}

		var errorMsg string
		if cfg.IsMultiDB {
			backupErr := errorHandler.HandleCombinedBackupError(
				cfg.OutputPath,
				err,
				stderrDetail,
				map[string]interface{}{
					"type": "combined_backup",
					"file": cfg.OutputPath,
				},
			)
			errorMsg = backupErr.Error()
		} else {
			errorMsg = errorHandler.HandleDatabaseBackupError(
				cfg.OutputPath,
				cfg.DBName,
				err,
				stderrDetail,
				map[string]interface{}{
					"database": cfg.DBName,
					"type":     cfg.BackupType + "_backup",
					"file":     cfg.OutputPath,
				},
			)
		}

		return types.DatabaseBackupInfo{}, fmt.Errorf("%s", errorMsg)
	}

	// Get file info and generate metadata
	backupInfo := s.buildBackupInfoFromResult(cfg, writeResult, timer, startTime)

	return backupInfo, nil
}

// buildBackupInfoFromResult membangun DatabaseBackupInfo dari result
func (s *Service) buildBackupInfoFromResult(cfg types_backup.BackupExecutionConfig, writeResult *types_backup.BackupWriteResult, timer *pkghelper.Timer, startTime time.Time) types.DatabaseBackupInfo {
	fileSize := writeResult.FileSize
	stderrOutput := writeResult.StderrOutput
	status := "success"

	if stderrOutput != "" {
		status = "success_with_warnings"
		if !cfg.IsMultiDB {
			s.Log.Warningf("Database %s backup dengan warning: %s", cfg.DBName, stderrOutput)
		}
	} else if !cfg.IsMultiDB {
		s.Log.Infof("✓ Database %s berhasil di-backup", cfg.DBName)
	}

	backupDuration := timer.Elapsed()
	endTime := time.Now()

	// Generate metadata
	var dbNames []string
	if cfg.IsMultiDB {
		dbNames = cfg.DBList
	} else {
		dbNames = []string{cfg.DBName}
	}

	meta := metadata.GenerateBackupMetadata(types_backup.MetadataConfig{
		BackupFile:      cfg.OutputPath,
		BackupType:      cfg.BackupType,
		DatabaseNames:   dbNames,
		Hostname:        s.BackupDBOptions.Profile.DBInfo.HostName,
		FileSize:        fileSize,
		Compressed:      s.BackupDBOptions.Compression.Enabled,
		CompressionType: s.BackupDBOptions.Compression.Type,
		Encrypted:       s.BackupDBOptions.Encryption.Enabled,
		BackupStatus:    status,
		StderrOutput:    stderrOutput,
		Duration:        backupDuration,
		StartTime:       startTime,
		EndTime:         endTime,
		Logger:          s.Log,
	})

	// Save metadata if configured
	manifestPath := ""
	if s.Config.Backup.Output.CreateBackupInfo || s.Config.Backup.Output.SaveBackupInfo {
		manifestPath = metadata.TrySaveBackupMetadata(meta, s.Log)
	}

	// Build backup info
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

// buildMysqldumpArgsFromConfig membangun mysqldump args dari konfigurasi backup
func (s *Service) buildMysqldumpArgsFromConfig(cfg types_backup.BackupExecutionConfig, profile types.ProfileInfo, filter types.FilterOptions, baseDumpArgs string) []string {
	conn := backuphelper.DatabaseConn{
		Host:     profile.DBInfo.Host,
		Port:     profile.DBInfo.Port,
		User:     profile.DBInfo.User,
		Password: profile.DBInfo.Password,
	}

	filterOpts := backuphelper.FilterOptions{
		ExcludeData:      filter.ExcludeData,
		ExcludeDatabases: filter.ExcludeDatabases,
		IncludeDatabases: filter.IncludeDatabases,
		ExcludeSystem:    filter.ExcludeSystem,
		ExcludeDBFile:    filter.ExcludeDBFile,
		IncludeFile:      filter.IncludeFile,
	}

	var dbList []string
	var singleDB string

	if cfg.IsMultiDB {
		dbList = cfg.DBList
	} else {
		singleDB = cfg.DBName
	}

	return backuphelper.BuildMysqldumpArgs(baseDumpArgs, conn, filterOpts, dbList, singleDB, cfg.TotalDBFound)
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

	s.Log.Info("Menyimpan informasi GTID ke file...")
	if err := metadata.SaveGTIDToFile(s.Log, gtidInfo, backupFilePath); err != nil {
		s.Log.Errorf("Gagal menyimpan GTID ke file: %v", err)
		return nil
	}

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
func (s *Service) exportUserGrantsIfNeeded(ctx context.Context, referenceBackupFile string) {
	if s.BackupDBOptions.ExcludeUser {
		return
	}

	if referenceBackupFile == "" {
		s.Log.Warn("Tidak ada backup file untuk export user grants")
		return
	}

	s.Log.Info("Export user grants ke file...")
	if err := metadata.ExportAndSaveUserGrants(ctx, s.Client, s.Log, referenceBackupFile, s.BackupDBOptions.ExcludeUser); err != nil {
		s.Log.Errorf("Gagal export user grants: %v", err)
	}
}

// =============================================================================
// Database Selection Helpers
// =============================================================================

// filterCandidatesByMode memfilter database candidates berdasarkan backup mode
func (s *Service) filterCandidatesByMode(dbFiltered []string, mode string) []string {
	candidates := make([]string, 0, len(dbFiltered))

	for _, db := range dbFiltered {
		dbLower := strings.ToLower(db)

		switch mode {
		case "primary":
			if strings.Contains(dbLower, "_secondary") || strings.HasSuffix(dbLower, "_dmart") ||
				strings.HasSuffix(dbLower, "_temp") || strings.HasSuffix(dbLower, "_archive") {
				continue
			}
		case "secondary":
			if !strings.Contains(dbLower, "_secondary") || strings.HasSuffix(dbLower, "_dmart") ||
				strings.HasSuffix(dbLower, "_temp") || strings.HasSuffix(dbLower, "_archive") {
				continue
			}
		case "single":
			if strings.HasSuffix(dbLower, "_dmart") || strings.HasSuffix(dbLower, "_temp") ||
				strings.HasSuffix(dbLower, "_archive") {
				continue
			}
		}

		candidates = append(candidates, db)
	}

	return candidates
}

// addCompanionDatabase menambahkan single companion database dengan suffix
func (s *Service) addCompanionDatabase(selectedDB string, suffix string, companionDbs *[]string,
	companionStatus map[string]bool, allDatabases []string) bool {

	dbName := selectedDB + suffix
	if pkghelper.StringSliceContainsFold(allDatabases, dbName) {
		s.Log.Infof("Menambahkan database companion: %s", dbName)
		*companionDbs = append(*companionDbs, dbName)
		companionStatus[dbName] = true
		return true
	} else {
		s.Log.Warnf("Database %s tidak ditemukan, melewati", dbName)
		companionStatus[dbName] = false
		return false
	}
}

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
		candidates := s.filterCandidatesByMode(dbFiltered, mode)

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

	if mode != "secondary" {
		if s.BackupDBOptions.IncludeDmart {
			s.addCompanionDatabase(selectedDB, "_dmart", &companionDbs, companionStatus, allDatabases)
		}
		if s.BackupDBOptions.IncludeTemp {
			s.addCompanionDatabase(selectedDB, "_temp", &companionDbs, companionStatus, allDatabases)
		}
		if s.BackupDBOptions.IncludeArchive {
			s.addCompanionDatabase(selectedDB, "_archive", &companionDbs, companionStatus, allDatabases)
		}
	}

	return companionDbs, selectedDB, companionStatus, nil
}
