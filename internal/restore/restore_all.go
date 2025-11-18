// File : internal/restore/restore_all.go
// Deskripsi : Restore semua database dari file combined backup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package restore

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/servicehelper"
	"sfDBTools/pkg/ui"
	"strings"
)

// executeRestoreAll melakukan restore semua database dari file combined backup
func (s *Service) executeRestoreAll(ctx context.Context) (types.RestoreResult, error) {
	defer servicehelper.TrackProgress(s)()

	// Mark restore as in progress untuk graceful shutdown
	s.SetRestoreInProgress(true)
	defer s.SetRestoreInProgress(false)

	timer := helper.NewTimer()
	sourceFile := s.RestoreOptions.SourceFile

	s.Log.Infof("Restoring all databases from: %s", filepath.Base(sourceFile))

	// Untuk combined backup, kita tidak tahu pasti database mana yang ada di dalam file
	// karena tidak ada metadata yang di-generate saat backup.
	// Combined backup akan me-restore semua database yang ada dalam file backup.
	// List databases akan kosong, yang artinya restore semua.
	databases := []string{} // Empty list = restore all databases from backup file

	// Check if dry run
	if s.RestoreOptions.DryRun {
		ui.PrintSubHeader("Mode Simulasi (Dry Run)")
		s.Log.Info("[DRY RUN] Restore tidak dijalankan (mode simulasi)")
		return BuildDryRunResult(databases, sourceFile), nil
	}

	// Pre-backup before restore (safety backup) - untuk combined backup
	// Karena kita tidak tahu database mana yang ada dalam backup, skip pre-backup
	// atau user harus manual backup sebelumnya
	var preBackupResult string
	if !s.RestoreOptions.SkipBackup && len(databases) > 0 {
		ui.PrintSubHeader("Membuat Safety Backup untuk Semua Database")
		s.Log.Infof("Creating safety backups untuk %d database sebelum restore...", len(databases))
		var preBackupFiles []string
		for _, dbName := range databases {
			preBackupFile, err := s.executePreBackup(ctx, dbName)
			if err != nil {
				s.Log.Warnf("Gagal create pre-backup untuk %s: %v", dbName, err)
				// Continue dengan backup database lainnya
				continue
			}
			preBackupFiles = append(preBackupFiles, preBackupFile)
		}
		if len(preBackupFiles) > 0 {
			preBackupResult = strings.Join(preBackupFiles, ", ")
			s.Log.Infof("✓ Safety backups created: %d files", len(preBackupFiles))
		}
	}

	// Execute restore all databases
	ui.PrintSubHeader("Melakukan Restore All Databases")
	s.Log.Info("Starting restore semua database dari combined backup...")
	restoreInfo, err := s.restoreAllDatabases(ctx, sourceFile, databases)
	duration := timer.Elapsed()

	// Build result dengan builder pattern
	builder := NewRestoreResultBuilder()
	builder.SetTotalDatabases(len(databases))
	if len(databases) == 0 {
		builder.SetTotalDatabases(1) // At least 1 untuk combined backup
	}
	if preBackupResult != "" {
		builder.SetPreBackupFile(preBackupResult)
	}

	if err != nil {
		builder.AddFailure("Combined", restoreInfo, duration, err)
	} else {
		builder.AddSuccess(restoreInfo, duration)
	}

	return builder.Build(), nil
}

// restoreAllDatabases melakukan restore semua database dari combined backup file
func (s *Service) restoreAllDatabases(ctx context.Context, sourceFile string, databases []string) (types.DatabaseRestoreInfo, error) {
	info := types.DatabaseRestoreInfo{
		DatabaseName:   fmt.Sprintf("Combined (%d databases)", len(databases)),
		SourceFile:     sourceFile,
		TargetDatabase: "Multiple databases",
	}

	// Get file info
	info.FileSize, info.FileSizeHuman = getFileInfo(sourceFile)

	// Setup reader pipeline: file → decrypt → decompress
	pipeline, err := s.setupReaderPipeline(sourceFile)
	if err != nil {
		return info, err
	}
	defer closePipeline(pipeline)

	// Execute mysql restore untuk semua database sekaligus
	// Combined backup sudah contain CREATE DATABASE statements
	if err := s.executeMysqlRestoreAll(ctx, pipeline.Reader); err != nil {
		return info, fmt.Errorf("gagal restore databases: %w", err)
	}

	info.Verified = true
	s.Log.Infof("✓ All databases berhasil di-restore")

	return info, nil
}

// executeMysqlRestoreAll menjalankan mysql command untuk restore all databases
func (s *Service) executeMysqlRestoreAll(ctx context.Context, reader io.Reader) error {
	opts := MysqlRestoreOptions{
		Host:           s.TargetProfile.DBInfo.Host,
		Port:           s.TargetProfile.DBInfo.Port,
		User:           s.TargetProfile.DBInfo.User,
		Password:       s.TargetProfile.DBInfo.Password,
		TargetDatabase: "", // Kosong untuk combined backup (sudah ada CREATE DATABASE di dalam)
		Force:          false,
		WithSpinner:    true,
	}

	return s.executeMysqlCommand(ctx, reader, opts)
}
