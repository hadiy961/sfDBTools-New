// File : internal/restore/service.go
// Deskripsi : Service utama untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 5 Januari 2026
package restore

import (
	restoremodel "sfDBTools/internal/app/restore/model"
	"sfDBTools/internal/app/restore/modes"
	"sfDBTools/internal/domain"
	appconfig "sfDBTools/internal/services/config"
	applog "sfDBTools/internal/services/log"
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
	Profile              *domain.ProfileInfo
	RestoreOpts          *restoremodel.RestoreSingleOptions
	RestorePrimaryOpts   *restoremodel.RestorePrimaryOptions
	RestoreSecondaryOpts *restoremodel.RestoreSecondaryOptions
	RestoreAllOpts       *restoremodel.RestoreAllOptions
	RestoreSelOpts       *restoremodel.RestoreSelectionOptions
	RestoreCustomOpts    *restoremodel.RestoreCustomOptions
	TargetClient         *database.Client

	// Restore-specific state
	restoreInProgress bool
	currentTargetDB   string
}

// NewRestoreService membuat instance baru Service dengan generic options
// Accepts: *restoremodel.RestoreSingleOptions, *restoremodel.RestorePrimaryOptions, *restoremodel.RestoreAllOptions, *restoremodel.RestoreSelectionOptions
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
		case *restoremodel.RestoreSingleOptions:
			svc.RestoreOpts = v
			svc.Profile = &v.Profile
		case *restoremodel.RestorePrimaryOptions:
			svc.RestorePrimaryOpts = v
			svc.Profile = &v.Profile
		case *restoremodel.RestoreSecondaryOptions:
			svc.RestoreSecondaryOpts = v
			svc.Profile = &v.Profile
		case *restoremodel.RestoreAllOptions:
			svc.RestoreAllOpts = v
			svc.Profile = &v.Profile
		case *restoremodel.RestoreSelectionOptions:
			svc.RestoreSelOpts = v
			svc.Profile = &v.Profile
		case *restoremodel.RestoreCustomOptions:
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

func (s *Service) GetProfile() *domain.ProfileInfo {
	return s.Profile
}

func (s *Service) GetSingleOptions() *restoremodel.RestoreSingleOptions {
	return s.RestoreOpts
}

func (s *Service) GetPrimaryOptions() *restoremodel.RestorePrimaryOptions {
	return s.RestorePrimaryOpts
}

func (s *Service) GetSecondaryOptions() *restoremodel.RestoreSecondaryOptions {
	return s.RestoreSecondaryOpts
}

func (s *Service) GetAllOptions() *restoremodel.RestoreAllOptions {
	return s.RestoreAllOpts
}

func (s *Service) GetSelectionOptions() *restoremodel.RestoreSelectionOptions {
	return s.RestoreSelOpts
}

func (s *Service) GetCustomOptions() *restoremodel.RestoreCustomOptions {
	return s.RestoreCustomOpts
}

// Ensure Service implements modes.RestoreService
var _ modes.RestoreService = (*Service)(nil)

// Other methods (BackupDatabaseIfNeeded, etc.) are implemented in service_helpers.go/restore_helpers.go
