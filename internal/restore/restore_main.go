package restore

// File : internal/restore/restore_main.go
// Deskripsi : Service layer untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

import (
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/errorlog"
	"sfDBTools/pkg/servicehelper"
)

// Service menyediakan operasi restore untuk database
type Service struct {
	servicehelper.BaseService // Embed base service untuk graceful shutdown functionality

	Config         *appconfig.Config
	Log            applog.Logger
	TargetProfile  *types.ProfileInfo
	RestoreOptions *types.RestoreOptions
	RestoreEntry   *types.RestoreEntryConfig
	Client         *database.Client // Client ke database target
	ErrorLog       *errorlog.ErrorLogger

	// Restore-specific state
	restoreInProgress bool
}

// NewRestoreService membuat instance Service baru untuk restore operations
func NewRestoreService(logs applog.Logger, cfg *appconfig.Config, restoreOpts interface{}) *Service {
	logDir := cfg.Log.Output.File.Dir
	if logDir == "" {
		logDir = "/var/log/sfDBTools"
	}

	svc := &Service{
		Log:      logs,
		Config:   cfg,
		ErrorLog: errorlog.NewErrorLogger(logs, logDir, "restore"),
	}

	if restoreOpts != nil {
		switch v := restoreOpts.(type) {
		case *types.RestoreOptions:
			svc.RestoreOptions = v
		default:
			logs.Warn("Tipe restore options tidak dikenali dalam Service")
		}
	} else {
		logs.Warn("Restore options tidak tersedia dalam Service")
	}

	return svc
}

// HandleShutdown menangani graceful shutdown saat restore
func (s *Service) HandleShutdown() {
	s.WithLock(func() {
		if s.restoreInProgress {
			s.Log.Warn("⚠ Restore sedang berlangsung, menghentikan...")
			// Cancel akan dipanggil via BaseService.Cancel()
		}
	})
	s.Cancel() // Panggil cancel dari BaseService
}

// SetRestoreInProgress menandai status restore sedang berlangsung
func (s *Service) SetRestoreInProgress(inProgress bool) {
	s.WithLock(func() {
		s.restoreInProgress = inProgress
	})
}
