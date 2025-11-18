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
	"sfDBTools/pkg/ui"
)

// executeRestoreSingle melakukan restore database dari satu file backup terpisah
func (s *Service) executeRestoreSingle(ctx context.Context) (types.RestoreResult, error) {
	defer servicehelper.TrackProgress(s)()

	// Mark restore as in progress untuk graceful shutdown
	s.SetRestoreInProgress(true)
	defer s.SetRestoreInProgress(false)

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
		ui.PrintSubHeader("Mode Simulasi (Dry Run)")
		s.Log.Info("[DRY RUN] Restore tidak dijalankan (mode simulasi)")
		dbName := sourceDatabaseName
		if dbName == "" {
			dbName = targetDB
		}
		return BuildDryRunResult([]string{dbName}, sourceFile), nil
	}

	// Prepare database: check exists → drop (jika flag aktif) → create (jika tidak ada)
	// Hasil preparation akan berisi info database existence SEBELUM preparation
	ui.PrintSubHeader("Mempersiapkan Database Target")
	prepResult, err := s.prepareDatabaseForRestore(ctx, targetDB, s.RestoreOptions.DropTarget)
	if err != nil {
		return types.RestoreResult{}, fmt.Errorf("gagal prepare database: %w", err)
	}

	// Pre-backup before restore (safety backup)
	// Gunakan info database existence dari prepResult untuk avoid duplicate check
	ui.PrintSubHeader("Membuat Safety Backup")
	preBackupFile, _, err := s.executePreBackupIfNeeded(ctx, targetDB, prepResult.DatabaseExists)
	if err != nil {
		return types.RestoreResult{}, fmt.Errorf("gagal create pre-backup: %w", err)
	}

	// Healthcheck koneksi sebelum restore
	s.Log.Debug("Validasi koneksi database sebelum restore...")
	if err := s.ensureValidConnection(ctx); err != nil {
		return types.RestoreResult{}, fmt.Errorf("gagal validasi koneksi database: %w", err)
	}

	// Execute restore
	ui.PrintSubHeader("Melakukan Restore Database")
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
