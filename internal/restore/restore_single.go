// File : internal/restore/restore_single.go
// Deskripsi : Restore database dari satu file backup terpisah
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package restore

import (
	"context"
	"fmt"
	"io"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/servicehelper"
)

// executeRestoreSingle melakukan restore database dari satu file backup terpisah
func (s *Service) executeRestoreSingle(ctx context.Context) (types.RestoreResult, error) {
	defer servicehelper.TrackProgress(s)()

	timer := helper.NewTimer()
	sourceFile := s.RestoreOptions.SourceFile

	// Resolve database name dengan priority: flag → metadata → filename → prompt
	dbNameResolution, err := s.resolveDatabaseName(sourceFile, s.RestoreOptions.TargetDB)
	if err != nil {
		return types.RestoreResult{}, err
	}

	targetDB := dbNameResolution.TargetDB
	sourceDatabaseName := dbNameResolution.SourceDB

	s.Log.Infof("Restore database: %s → %s (resolved from: %s)",
		sourceDatabaseName, targetDB, dbNameResolution.ResolvedFrom)

	// Check if dry run
	if s.RestoreOptions.DryRun {
		s.Log.Info("[DRY RUN] Restore tidak dijalankan (mode simulasi)")
		dbName := sourceDatabaseName
		if dbName == "" {
			dbName = targetDB
		}
		return BuildDryRunResult([]string{dbName}, sourceFile), nil
	}

	// Pre-backup before restore (safety backup)
	preBackupFile, _, err := s.executePreBackupIfNeeded(ctx, targetDB)
	if err != nil {
		return types.RestoreResult{}, fmt.Errorf("gagal create pre-backup: %w", err)
	}

	// Prepare database: drop (jika flag aktif) → create (jika tidak ada)
	if _, err := s.prepareDatabaseForRestore(ctx, targetDB, s.RestoreOptions.DropTarget); err != nil {
		return types.RestoreResult{}, fmt.Errorf("gagal prepare database: %w", err)
	}

	// Execute restore
	restoreInfo, err := s.restoreSingleDatabase(ctx, sourceFile, targetDB, sourceDatabaseName)
	duration := timer.Elapsed()

	// Build result dengan builder pattern
	if err != nil {
		result := BuildSingleDBFailureResult(targetDB, restoreInfo, duration, err)
		if preBackupFile != "" {
			result.PreBackupFile = preBackupFile
		}
		return result, nil
	}

	// Success - check if ada warning
	if restoreInfo.Warnings != "" {
		s.Log.Warnf("Restore success dengan warning: %s", restoreInfo.Warnings)
		result := BuildSingleDBSuccessResult(restoreInfo, duration, restoreInfo.Warnings)
		if preBackupFile != "" {
			result.PreBackupFile = preBackupFile
		}
		return result, nil
	}

	// Success tanpa warning
	result := BuildSingleDBSuccessResult(restoreInfo, duration, "")
	if preBackupFile != "" {
		result.PreBackupFile = preBackupFile
	}
	return result, nil
}

// restoreSingleDatabase melakukan restore satu database dari backup file
func (s *Service) restoreSingleDatabase(ctx context.Context, sourceFile, targetDB, sourceDatabaseName string) (types.DatabaseRestoreInfo, error) {
	dbName := sourceDatabaseName
	info := types.DatabaseRestoreInfo{
		DatabaseName:   dbName,
		SourceFile:     sourceFile,
		TargetDatabase: targetDB,
	}

	// Get file info
	info.FileSize, info.FileSizeHuman = getFileInfo(sourceFile)

	// Setup reader pipeline: file → decrypt → decompress
	pipeline, err := s.setupReaderPipeline(sourceFile)
	if err != nil {
		return info, err
	}
	defer closePipeline(pipeline)

	// Setup max_statement_time untuk GLOBAL restore (set ke unlimited untuk restore jangka panjang)
	restoreMaxStmt, err := s.setupMaxStatementTimeForRestore(ctx)
	if err == nil {
		defer func() {
			if rerr := restoreMaxStmt(context.Background()); rerr != nil {
				// Error sudah di-log di dalam function
			}
		}()
	}

	// Check max_allowed_packet sebelum restore
	s.checkMaxAllowedPacket(ctx)

	// Execute mysql restore menggunakan pipe
	if err := s.executeMysqlRestore(ctx, pipeline.Reader, targetDB, sourceFile, sourceDatabaseName); err != nil {
		s.Log.Debugf("[DEBUG] MySQL restore error detected: %v, Force=%v", err, s.RestoreOptions.Force)
		// Jika force=true, log warning tapi tetap success (dengan warning)
		// Jika force=false, return error (restore gagal)
		if s.RestoreOptions.Force {
			s.Log.Warnf("MySQL restore memiliki error tapi tetap berjalan (--force mode): %v", err)
			info.Warnings = fmt.Sprintf("MySQL restore memiliki error tapi tetap berjalan: %v", err)
		} else {
			s.Log.Errorf("[ERROR] Restore gagal dengan force=false, returning error: %v", err)
			return info, fmt.Errorf("gagal restore database: %w", err)
		}
	} else {
		s.Log.Debugf("[DEBUG] MySQL restore completed without error")
	}

	info.Verified = true
	s.Log.Infof("✓ Database %s berhasil di-restore", targetDB)

	return info, nil
}

// executeMysqlRestore menjalankan mysql command untuk restore dari reader
func (s *Service) executeMysqlRestore(ctx context.Context, reader io.Reader, targetDB, sourceFile, sourceDatabaseName string) error {
	opts := MysqlRestoreOptions{
		Host:           s.TargetProfile.DBInfo.Host,
		Port:           s.TargetProfile.DBInfo.Port,
		User:           s.TargetProfile.DBInfo.User,
		Password:       s.TargetProfile.DBInfo.Password,
		TargetDatabase: targetDB,
		Force:          s.RestoreOptions.Force,
		WithSpinner:    true,
	}

	err := s.executeMysqlCommand(ctx, reader, opts)
	if err != nil {
		s.Log.Error("MySQL restore gagal, lihat log untuk detail")
		// Log error detail ke file terpisah
		s.logErrorWithDetail(map[string]interface{}{
			"database": sourceDatabaseName,
			"source":   sourceFile,
			"target":   targetDB,
			"force":    s.RestoreOptions.Force,
		}, err.Error(), err)
		// Return error, logic force handling di level caller
		return err
	}

	return nil
}
