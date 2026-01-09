// File : internal/app/backup/service.go
// Deskripsi : Service utama untuk backup operations dengan interface implementation
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2026-01-05
package backup

import (
	"fmt"
	"sfdbtools/internal/app/backup/gtid"
	"sfdbtools/internal/app/backup/helpers/compression"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/app/backup/modes"
	"sfdbtools/internal/domain"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/shared/errorlog"
	"sfdbtools/internal/shared/fsops"
	"sfdbtools/internal/shared/servicehelper"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/progress"
)

// Service adalah service utama untuk backup operations
type Service struct {
	servicehelper.BaseService

	Config          *appconfig.Config
	Log             applog.Logger
	ErrorLog        *errorlog.ErrorLogger
	DBInfo          *domain.DBInfo
	Profile         *domain.ProfileInfo
	BackupDBOptions *types_backup.BackupDBOptions
	BackupEntry     *types_backup.BackupEntryConfig
	Client          *database.Client

	gtidInfo          *gtid.GTIDInfo
	currentBackupFile string
	backupInProgress  bool
	excludedDatabases []string // List database yang dikecualikan (untuk mode 'all')
}

// NewBackupService membuat instance baru Service
func NewBackupService(logs applog.Logger, cfg *appconfig.Config, backup interface{}) *Service {
	logDir := cfg.Log.Output.File.Dir
	if logDir == "" {
		logDir = consts.DefaultLogDir
	}

	svc := &Service{
		Log:               logs,
		Config:            cfg,
		ErrorLog:          errorlog.NewErrorLogger(logs, logDir, consts.FeatureBackup),
		excludedDatabases: []string{}, // Initialize dengan empty slice, bukan nil
	}

	if backup != nil {
		switch v := backup.(type) {
		case *types_backup.BackupDBOptions:
			svc.BackupDBOptions = v
			svc.Profile = &v.Profile
			svc.BackupEntry = &v.Entry
			svc.DBInfo = &v.Profile.DBInfo
		default:
			logs.Warn("Tipe backup tidak dikenali dalam Service")
		}
	} else {
		logs.Warn("Tipe backup tidak dikenali dalam Service")
	}

	return svc
}

// Verify interface implementation at compile time
var _ modes.BackupService = (*Service)(nil)

// buildCompressionSettings delegates ke shared helper
func (s *Service) buildCompressionSettings() types_backup.CompressionSettings {
	return compression.BuildCompressionSettings(s.BackupDBOptions)
}

// =============================================================================
// Interface helpers (used by modes.BackupService)
// =============================================================================

// GetLog returns logger instance
func (s *Service) GetLog() applog.Logger { return s.Log }

// GetOptions returns backup options
func (s *Service) GetOptions() *types_backup.BackupDBOptions { return s.BackupDBOptions }

// ToBackupResult konversi BackupLoopResult ke BackupResult
func (s *Service) ToBackupResult(loopResult types_backup.BackupLoopResult) types_backup.BackupResult {
	return types_backup.BackupResult{
		BackupInfo:          loopResult.BackupInfos,
		FailedDatabaseInfos: loopResult.FailedDBs,
		Errors:              loopResult.Errors,
	}
}

// =============================================================================
// Service state / shutdown
// =============================================================================

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
			shouldRemoveFile = true
			fileToRemove = s.currentBackupFile
			s.currentBackupFile = ""
			s.backupInProgress = false
		}
	})

	if shouldRemoveFile {
		progress.RunWithSpinnerSuspended(func() {
			s.Log.Warn("Proses backup dihentikan, melakukan rollback...")
			if err := fsops.RemoveFile(fileToRemove); err != nil {
				s.Log.Errorf("Gagal menghapus file backup: %v", err)
				print.PrintError(fmt.Sprintf("⚠ WARNING: File backup partial mungkin masih tersisa: %s", fileToRemove))
				print.PrintError("Silakan hapus manual jika diperlukan.")
			} else {
				s.Log.Infof("File backup yang belum selesai berhasil dihapus: %s", fileToRemove)
				s.Log.Info("File backup partial berhasil dihapus")
			}
		})
	} else {
		progress.RunWithSpinnerSuspended(func() {
			s.Log.Warn("Menerima signal interrupt, tidak ada proses backup aktif")
		})
	}

	s.Cancel()
}
