// File : internal/restore/service.go
// Deskripsi : Service utama untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-17

package restore

import (
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/restore/modes"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/errorlog"
	"sfDBTools/pkg/servicehelper"
)

// Service adalah service utama untuk restore operations
type Service struct {
	servicehelper.BaseService

	Config             *appconfig.Config
	Log                applog.Logger
	ErrorLog           *errorlog.ErrorLogger
	Profile            *types.ProfileInfo
	RestoreOpts        *types.RestoreSingleOptions
	RestorePrimaryOpts *types.RestorePrimaryOptions
	RestoreAllOpts     *types.RestoreAllOptions
	TargetClient       *database.Client

	// Restore-specific state
	restoreInProgress bool
	currentTargetDB   string
}

// NewRestoreService membuat instance baru Service untuk single mode
func NewRestoreService(logs applog.Logger, cfg *appconfig.Config, opts *types.RestoreSingleOptions) *Service {
	logDir := cfg.Log.Output.File.Dir
	if logDir == "" {
		logDir = "/var/log/sfDBTools"
	}

	return &Service{
		Log:         logs,
		Config:      cfg,
		ErrorLog:    errorlog.NewErrorLogger(logs, logDir, "restore"),
		RestoreOpts: opts,
		Profile:     &opts.Profile,
	}
}

// NewRestorePrimaryService membuat instance baru Service untuk primary mode
func NewRestorePrimaryService(logs applog.Logger, cfg *appconfig.Config, opts *types.RestorePrimaryOptions) *Service {
	logDir := cfg.Log.Output.File.Dir
	if logDir == "" {
		logDir = "/var/log/sfDBTools"
	}

	return &Service{
		Log:                logs,
		Config:             cfg,
		ErrorLog:           errorlog.NewErrorLogger(logs, logDir, "restore"),
		RestorePrimaryOpts: opts,
		Profile:            &opts.Profile,
	}
}

// NewRestoreAllService membuat instance baru Service untuk all mode
func NewRestoreAllService(logs applog.Logger, cfg *appconfig.Config, opts *types.RestoreAllOptions) *Service {
	logDir := cfg.Log.Output.File.Dir
	if logDir == "" {
		logDir = "/var/log/sfDBTools"
	}

	return &Service{
		Log:            logs,
		Config:         cfg,
		ErrorLog:       errorlog.NewErrorLogger(logs, logDir, "restore"),
		RestoreAllOpts: opts,
		Profile:        &opts.Profile,
	}
}

// SetRestoreInProgress mencatat status restore yang sedang berjalan
func (s *Service) SetRestoreInProgress(dbName string) {
	s.WithLock(func() {
		s.currentTargetDB = dbName
		s.restoreInProgress = true
	})
}

// ClearRestoreInProgress menghapus catatan restore setelah selesai
func (s *Service) ClearRestoreInProgress() {
	s.WithLock(func() {
		s.currentTargetDB = ""
		s.restoreInProgress = false
	})
}

// HandleShutdown menangani graceful shutdown saat CTRL+C atau interrupt
func (s *Service) HandleShutdown() {
	s.WithLock(func() {
		if s.restoreInProgress {
			s.Log.Warn("Proses restore dihentikan oleh user, database mungkin dalam keadaan tidak konsisten")
			s.restoreInProgress = false
			s.currentTargetDB = ""
		}
	})

	// Close database connection
	if s.TargetClient != nil {
		s.TargetClient.Close()
	}

	s.Log.Info("Restore service shutdown completed")
}

// Close cleanup resources
func (s *Service) Close() error {
	if s.TargetClient != nil {
		return s.TargetClient.Close()
	}
	return nil
}

// =============================================================================
// Interface Implementation - modes.RestoreService
// =============================================================================

func (s *Service) GetLogger() applog.Logger {
	return s.Log
}

func (s *Service) LogInfo(msg string) {
	s.Log.Info(msg)
}

func (s *Service) LogWarn(msg string) {
	s.Log.Warn(msg)
}

func (s *Service) LogWarnf(format string, args ...interface{}) {
	s.Log.Warnf(format, args...)
}

func (s *Service) LogInfof(format string, args ...interface{}) {
	s.Log.Infof(format, args...)
}

func (s *Service) LogDebugf(format string, args ...interface{}) {
	s.Log.Debugf(format, args...)
}

func (s *Service) LogError(msg string) {
	s.Log.Error(msg)
}

func (s *Service) LogErrorf(format string, args ...interface{}) {
	s.Log.Errorf(format, args...)
}

func (s *Service) GetTargetClient() *database.Client {
	return s.TargetClient
}

func (s *Service) GetProfile() *types.ProfileInfo {
	return s.Profile
}

func (s *Service) GetSingleOptions() *types.RestoreSingleOptions {
	return s.RestoreOpts
}

func (s *Service) GetPrimaryOptions() *types.RestorePrimaryOptions {
	return s.RestorePrimaryOpts
}

func (s *Service) GetAllOptions() *types.RestoreAllOptions {
	return s.RestoreAllOpts
}

// Ensure Service implements modes.RestoreService
var _ modes.RestoreService = (*Service)(nil)

// Other methods (BackupDatabaseIfNeeded, etc.) are implemented in service_helpers.go/restore_helpers.go
