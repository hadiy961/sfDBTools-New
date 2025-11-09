package restore

// File : internal/restore/restore_main.go
// Deskripsi : Service layer untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

import (
	"context"
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sync"
)

// Service menyediakan operasi restore untuk database
type Service struct {
	Config         *appconfig.Config
	Log            applog.Logger
	TargetProfile  *types.ProfileInfo
	RestoreOptions *types.RestoreOptions
	RestoreEntry   *types.RestoreEntryConfig
	Client         *database.Client // Client ke database target

	// Untuk graceful shutdown
	cancelFunc        context.CancelFunc
	restoreInProgress bool
	mu                sync.Mutex
}

// NewRestoreService membuat instance Service baru untuk restore operations
func NewRestoreService(logs applog.Logger, cfg *appconfig.Config, restoreOpts interface{}) *Service {
	svc := &Service{
		Log:    logs,
		Config: cfg,
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

// SetCancelFunc mengatur context cancel function untuk graceful shutdown
func (s *Service) SetCancelFunc(cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancelFunc = cancel
}

// HandleShutdown menangani graceful shutdown saat restore
func (s *Service) HandleShutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.restoreInProgress {
		s.Log.Warn("⚠ Restore sedang berlangsung, menghentikan...")
		if s.cancelFunc != nil {
			s.cancelFunc()
		}
	}
}

// SetRestoreInProgress menandai status restore sedang berlangsung
func (s *Service) SetRestoreInProgress(inProgress bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.restoreInProgress = inProgress
}
