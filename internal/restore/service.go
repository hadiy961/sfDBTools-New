// File : internal/restore/service.go
// Deskripsi : Service utama untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified :  2026-01-05
package restore

import (
	"sfDBTools/internal/services/config"
	"sfDBTools/internal/services/log"
	"sfDBTools/internal/restore/modes"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/errorlog"
	"sfDBTools/pkg/servicehelper"
)

// Service adalah service utama untuk restore operations
type Service struct {
	servicehelper.BaseService

	Config               *appconfig.Config
	Log                  applog.Logger
	ErrorLog             *errorlog.ErrorLogger
	Profile              *types.ProfileInfo
	RestoreOpts          *types.RestoreSingleOptions
	RestorePrimaryOpts   *types.RestorePrimaryOptions
	RestoreSecondaryOpts *types.RestoreSecondaryOptions
	RestoreAllOpts       *types.RestoreAllOptions
	RestoreSelOpts       *types.RestoreSelectionOptions
	RestoreCustomOpts    *types.RestoreCustomOptions
	TargetClient         *database.Client

	// Restore-specific state
	restoreInProgress bool
	currentTargetDB   string
}

// NewRestoreService membuat instance baru Service dengan generic options
// Accepts: *types.RestoreSingleOptions, *types.RestorePrimaryOptions, *types.RestoreAllOptions, *types.RestoreSelectionOptions
func NewRestoreService(logs applog.Logger, cfg *appconfig.Config, restore interface{}) *Service {
	logDir := cfg.Log.Output.File.Dir
	if logDir == "" {
		logDir = consts.DefaultLogDir
	}

	svc := &Service{
		Log:      logs,
		Config:   cfg,
		ErrorLog: errorlog.NewErrorLogger(logs, logDir, consts.FeatureRestore),
	}

	if restore != nil {
		switch v := restore.(type) {
		case *types.RestoreSingleOptions:
			svc.RestoreOpts = v
			svc.Profile = &v.Profile
		case *types.RestorePrimaryOptions:
			svc.RestorePrimaryOpts = v
			svc.Profile = &v.Profile
		case *types.RestoreSecondaryOptions:
			svc.RestoreSecondaryOpts = v
			svc.Profile = &v.Profile
		case *types.RestoreAllOptions:
			svc.RestoreAllOpts = v
			svc.Profile = &v.Profile
		case *types.RestoreSelectionOptions:
			svc.RestoreSelOpts = v
			svc.Profile = &v.Profile
		case *types.RestoreCustomOptions:
			svc.RestoreCustomOpts = v
			svc.Profile = &v.Profile
		default:
			logs.Warn("Tipe restore options tidak dikenali dalam Service")
		}
	} else {
		logs.Warn("Restore options tidak diberikan ke Service")
	}

	return svc
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

func (s *Service) GetSecondaryOptions() *types.RestoreSecondaryOptions {
	return s.RestoreSecondaryOpts
}

func (s *Service) GetAllOptions() *types.RestoreAllOptions {
	return s.RestoreAllOpts
}

func (s *Service) GetSelectionOptions() *types.RestoreSelectionOptions {
	return s.RestoreSelOpts
}

func (s *Service) GetCustomOptions() *types.RestoreCustomOptions {
	return s.RestoreCustomOpts
}

// Ensure Service implements modes.RestoreService
var _ modes.RestoreService = (*Service)(nil)

// Other methods (BackupDatabaseIfNeeded, etc.) are implemented in service_helpers.go/restore_helpers.go
