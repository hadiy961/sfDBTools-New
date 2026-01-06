// File : internal/app/cleanup/service.go
// Deskripsi : Service utama implementation untuk cleanup operations
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 5 Januari 2026
package cleanup

import (
	"errors"

	cleanupmodel "sfdbtools/internal/app/cleanup/model"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
)

// Error definitions
var (
	ErrInvalidCleanupMode = errors.New("mode cleanup tidak valid")
)

// Service adalah service untuk cleanup operations
type Service struct {
	Config         *appconfig.Config
	Log            applog.Logger
	CleanupOptions cleanupmodel.CleanupOptions
}

// NewCleanupService membuat instance baru dari Service dengan proper dependency injection
func NewCleanupService(config *appconfig.Config, logger applog.Logger, opts cleanupmodel.CleanupOptions) *Service {
	return &Service{
		Config:         config,
		Log:            logger,
		CleanupOptions: opts,
	}
}

// ExecuteCleanupCommand adalah entry point utama untuk cleanup execution
func (s *Service) ExecuteCleanupCommand(config cleanupmodel.CleanupEntryConfig) error {
	// Log prefix untuk tracking
	if config.LogPrefix != "" {
		s.Log.Infof("[%s] Memulai cleanup dengan mode: %s", config.LogPrefix, config.Mode)
	}

	// Tampilkan options jika diminta
	if config.ShowOptions {
		s.displayCleanupOptions()
	}

	// Jalankan cleanup berdasarkan mode
	dryRun := s.CleanupOptions.DryRun
	switch config.Mode {
	case "run":
		return s.cleanupCore(dryRun, s.CleanupOptions.Pattern)
	case "pattern":
		if s.CleanupOptions.Pattern == "" {
			return ErrInvalidCleanupMode
		}
		return s.cleanupCore(dryRun, s.CleanupOptions.Pattern)
	default:
		return ErrInvalidCleanupMode
	}
}
