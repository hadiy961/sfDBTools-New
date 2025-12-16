// File : internal/backup/service.go
// Deskripsi : Service utama untuk backup operations dengan interface implementation
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package backup

import (
	"context"
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/backup/modes"
	"sfDBTools/internal/cleanup"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/errorlog"
	"sfDBTools/pkg/servicehelper"
)

// Service adalah service utama untuk backup operations
type Service struct {
	servicehelper.BaseService

	Config          *appconfig.Config
	Log             applog.Logger
	ErrorLog        *errorlog.ErrorLogger
	BackupDBOptions *types_backup.BackupDBOptions
	Client          *database.Client

	// Backup-specific state
	currentBackupFile string
	backupInProgress  bool
	gtidInfo          *database.GTIDInfo
	excludedDatabases []string // List database yang dikecualikan (untuk mode 'all')
}

// NewBackupService membuat instance baru Service
func NewBackupService(logs applog.Logger, cfg *appconfig.Config, backup interface{}) *Service {
	logDir := cfg.Log.Output.File.Dir
	if logDir == "" {
		logDir = "/var/log/sfDBTools"
	}

	svc := &Service{
		Log:               logs,
		Config:            cfg,
		ErrorLog:          errorlog.NewErrorLogger(logs, logDir, "backup"),
		excludedDatabases: []string{}, // Initialize dengan empty slice, bukan nil
	}

	if backup != nil {
		switch v := backup.(type) {
		case *types_backup.BackupDBOptions:
			svc.BackupDBOptions = v
		default:
			logs.Warn("Tipe backup tidak dikenali dalam Service")
		}
	} else {
		logs.Warn("Tipe backup tidak dikenali dalam Service")
	}

	return svc
}

// SetCurrentBackupFile mencatat file backup yang sedang dibuat
func (s *Service) SetCurrentBackupFile(filePath string) {
	s.WithLock(func() {
		s.currentBackupFile = filePath
		s.backupInProgress = true
	})
}

// ClearCurrentBackupFile menghapus catatan file backup setelah selesai
func (s *Service) ClearCurrentBackupFile() {
	s.WithLock(func() {
		s.currentBackupFile = ""
		s.backupInProgress = false
	})
}

// HandleShutdown menangani graceful shutdown saat CTRL+C atau interrupt
func (s *Service) HandleShutdown() {
	var shouldRemoveFile bool
	var fileToRemove string

	s.WithLock(func() {
		if s.backupInProgress && s.currentBackupFile != "" {
			s.Log.Warn("Proses backup dihentikan, melakukan rollback...")
			shouldRemoveFile = true
			fileToRemove = s.currentBackupFile
			s.currentBackupFile = ""
			s.backupInProgress = false
		}
	})

	if shouldRemoveFile {
		if err := cleanup.CleanupPartialBackup(fileToRemove, s.Log); err != nil {
			s.Log.Errorf("Gagal menghapus file backup: %v", err)
		}
	}

	s.Cancel()
}

// =============================================================================
// Interface Implementation - modes.BackupService
// =============================================================================

// GetLogger implements modes.BackupService
func (s *Service) GetLogger() applog.Logger {
	return s.Log
}

// GetBackupOptions implements modes.BackupService
func (s *Service) GetBackupOptions() *types_backup.BackupDBOptions {
	return s.BackupDBOptions
}

// ExecuteAndBuildBackup implements modes.BackupService
func (s *Service) ExecuteAndBuildBackup(ctx context.Context, cfg types_backup.BackupExecutionConfig) (types.DatabaseBackupInfo, error) {
	return s.executeAndBuildBackup(ctx, cfg)
}

// ExecuteBackupLoop implements modes.BackupService
func (s *Service) ExecuteBackupLoop(ctx context.Context, databases []string, config types_backup.BackupLoopConfig, outputPathFunc func(dbName string) (string, error)) types_backup.BackupLoopResult {
	return s.executeBackupLoop(ctx, databases, config, outputPathFunc)
}

// GenerateFullBackupPath implements modes.BackupService
func (s *Service) GenerateFullBackupPath(dbName string, mode string) (string, error) {
	return s.generateFullBackupPath(dbName, mode)
}

// GetTotalDatabaseCount implements modes.BackupService
func (s *Service) GetTotalDatabaseCount(ctx context.Context, dbFiltered []string) int {
	return s.getTotalDatabaseCount(ctx, dbFiltered)
}

// CaptureAndSaveGTID implements modes.BackupService
func (s *Service) CaptureAndSaveGTID(ctx context.Context, backupFilePath string) error {
	return s.captureAndSaveGTID(ctx, backupFilePath)
}

// ExportUserGrantsIfNeeded implements modes.BackupService
func (s *Service) ExportUserGrantsIfNeeded(ctx context.Context, referenceBackupFile string, databases []string) string {
	return s.exportUserGrantsIfNeeded(ctx, referenceBackupFile, databases)
}

// UpdateMetadataUserGrantsPath implements modes.BackupService
func (s *Service) UpdateMetadataUserGrantsPath(backupFilePath string, userGrantsPath string) {
	s.updateMetadataUserGrantsPath(backupFilePath, userGrantsPath)
}

// ToBackupResult implements modes.BackupService - konversi modes.BackupLoopResult ke types_backup.BackupResult
func (s *Service) ToBackupResult(loopResult types_backup.BackupLoopResult) types_backup.BackupResult {
	return types_backup.BackupResult{
		BackupInfo:          loopResult.BackupInfos,
		FailedDatabaseInfos: loopResult.FailedDBs,
		Errors:              loopResult.Errors,
	}
}

// Verify interface implementation at compile time
var _ modes.BackupService = (*Service)(nil)
