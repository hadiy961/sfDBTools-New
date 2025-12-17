// File : internal/cleanup/service.go
// Deskripsi : Service utama implementation untuk cleanup operations
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 17 Desember 2025

package cleanup

import (
	"errors"
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
)

// Error definitions
var (
	ErrInvalidCleanupMode = errors.New("mode cleanup tidak valid")
)

// Service adalah service untuk cleanup operations
type Service struct {
	Config         *appconfig.Config
	Log            applog.Logger
	CleanupOptions types.CleanupOptions
}

// NewCleanupService membuat instance baru dari Service dengan proper dependency injection
func NewCleanupService(config *appconfig.Config, logger applog.Logger, opts types.CleanupOptions) *Service {
	return &Service{
		Config:         config,
		Log:            logger,
		CleanupOptions: opts,
	}
}

// ExecuteCleanupCommand adalah entry point utama untuk cleanup execution
func (s *Service) ExecuteCleanupCommand(config types.CleanupEntryConfig) error {
	// Log prefix untuk tracking
	if config.LogPrefix != "" {
		s.Log.Infof("[%s] Memulai cleanup dengan mode: %s", config.LogPrefix, config.Mode)
	}

	// Tampilkan options jika diminta
	if config.ShowOptions {
		s.displayCleanupOptions()
	}

	// Jalankan cleanup berdasarkan mode
	switch config.Mode {
	case "run":
		return s.cleanupCore(false, s.CleanupOptions.Pattern)
	case "dry-run":
		return s.cleanupCore(true, s.CleanupOptions.Pattern)
	case "pattern":
		if s.CleanupOptions.Pattern == "" {
			return ErrInvalidCleanupMode
		}
		return s.cleanupCore(false, s.CleanupOptions.Pattern)
	default:
		return ErrInvalidCleanupMode
	}
}
