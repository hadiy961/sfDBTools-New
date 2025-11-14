// File : internal/restore/restore_database.go
// Deskripsi : Helper functions untuk database preparation dan management
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-11
// Last Modified : 2025-11-11

package restore

import (
	"context"
	"fmt"
	"sfDBTools/internal/types"
)

// DatabasePreparationResult berisi hasil dari persiapan database
type DatabasePreparationResult struct {
	DatabaseName      string
	DatabaseExists    bool   // Database exists sebelum preparation
	DatabaseCreated   bool   // Database baru dibuat
	DatabaseDropped   bool   // Database di-drop
	PreBackupExecuted bool   // Pre-backup dijalankan
	PreBackupFile     string // Path file pre-backup jika dijalankan
	Skipped           bool   // Preparation di-skip (e.g., database tidak ada dan tidak perlu backup)
	ErrorMessage      string // Error message jika gagal
}

// prepareDatabaseForRestore melakukan persiapan database sebelum restore
// Menangani: check exists → drop (jika diperlukan) → create (jika diperlukan)
// Returns: DatabasePreparationResult dan error
func (s *Service) prepareDatabaseForRestore(ctx context.Context, targetDB string, dropTarget bool) (*DatabasePreparationResult, error) {
	result := &DatabasePreparationResult{
		DatabaseName: targetDB,
	}

	// Pastikan koneksi valid sebelum operasi database
	if err := s.ensureValidConnection(ctx); err != nil {
		result.ErrorMessage = fmt.Sprintf("gagal ensure valid connection: %v", err)
		return result, fmt.Errorf("gagal ensure database connection: %w", err)
	}

	// Check apakah database target sudah ada
	exists, err := s.isTargetDatabaseExists(ctx, targetDB)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("gagal check keberadaan database: %v", err)
		return result, fmt.Errorf("gagal check database existence: %w", err)
	}

	result.DatabaseExists = exists

	// Jika flag --drop-target aktif dan database ada, drop database
	if dropTarget && exists {
		s.Log.Infof("Dropping target database: %s", targetDB)
		if err := s.Client.DropDatabase(ctx, targetDB); err != nil {
			result.ErrorMessage = fmt.Sprintf("gagal drop database: %v", err)
			return result, fmt.Errorf("gagal drop target database %s: %w", targetDB, err)
		}
		result.DatabaseDropped = true
		result.DatabaseExists = false // Database sudah di-drop
		s.Log.Infof("✓ Database %s berhasil di-drop", targetDB)
	}

	// Jika database tidak ada (either karena di-drop atau memang belum ada), buat database baru
	if !result.DatabaseExists {
		s.Log.Infof("Database target tidak ada, membuat database baru: %s", targetDB)
		if err := s.Client.CreateDatabase(ctx, targetDB); err != nil {
			result.ErrorMessage = fmt.Sprintf("gagal create database: %v", err)
			return result, fmt.Errorf("gagal create target database: %w", err)
		}
		result.DatabaseCreated = true
		s.Log.Infof("✓ Database %s berhasil dibuat", targetDB)
	}

	return result, nil
}

// executePreBackupIfNeeded menjalankan pre-backup jika diperlukan berdasarkan kondisi
// Kondisi yang di-check:
// - SkipBackup flag
// - Database existence
// - Dry run mode
// Returns: backup file path, executed status, error
func (s *Service) executePreBackupIfNeeded(ctx context.Context, targetDB string) (string, bool, error) {
	// Skip pre-backup jika flag --skip-backup aktif
	if s.RestoreOptions.SkipBackup {
		s.Log.Debug("Pre-backup skipped (--skip-backup flag aktif)")
		return "", false, nil
	}

	// Skip pre-backup jika dry run mode
	if s.RestoreOptions.DryRun {
		s.Log.Debug("Pre-backup skipped (dry run mode)")
		return "", false, nil
	}

	// Pastikan koneksi valid sebelum check database
	if err := s.ensureValidConnection(ctx); err != nil {
		return "", false, fmt.Errorf("gagal ensure valid connection: %w", err)
	}

	// Check apakah database target ada
	exists, err := s.isTargetDatabaseExists(ctx, targetDB)
	if err != nil {
		return "", false, fmt.Errorf("gagal check database existence untuk pre-backup: %w", err)
	}

	// Skip pre-backup jika database tidak ada (tidak ada yang perlu di-backup)
	if !exists {
		s.Log.Infof("Database target tidak ada, skip pre-backup untuk: %s", targetDB)
		return "", false, nil
	}

	// Execute pre-backup
	s.Log.Infof("Creating safety backup before restore untuk: %s", targetDB)
	preBackupFile, err := s.executePreBackup(ctx, targetDB)
	if err != nil {
		return "", false, fmt.Errorf("gagal create pre-backup: %w", err)
	}

	s.Log.Infof("✓ Safety backup created: %s", preBackupFile)
	return preBackupFile, true, nil
}

// buildRestoreInfo membuat DatabaseRestoreInfo object dengan informasi lengkap
func buildRestoreInfo(databaseName, sourceFile, targetDB, status string, fileSize int64, fileSizeHuman string) types.DatabaseRestoreInfo {
	return types.DatabaseRestoreInfo{
		DatabaseName:   databaseName,
		SourceFile:     sourceFile,
		TargetDatabase: targetDB,
		Status:         status,
		FileSize:       fileSize,
		FileSizeHuman:  fileSizeHuman,
	}
}

// buildFailedRestoreInfo membuat DatabaseRestoreInfo untuk restore yang gagal dengan error message
func buildFailedRestoreInfo(databaseName, sourceFile, targetDB, errorMsg string, fileSize int64, fileSizeHuman string) types.DatabaseRestoreInfo {
	info := buildRestoreInfo(databaseName, sourceFile, targetDB, "failed", fileSize, fileSizeHuman)
	info.ErrorMessage = errorMsg
	return info
}

// buildSkippedRestoreInfo membuat DatabaseRestoreInfo untuk restore yang di-skip (dry run)
func buildSkippedRestoreInfo(databaseName, sourceFile, targetDB string, fileSize int64, fileSizeHuman string) types.DatabaseRestoreInfo {
	return buildRestoreInfo(databaseName, sourceFile, targetDB, "skipped", fileSize, fileSizeHuman)
}
