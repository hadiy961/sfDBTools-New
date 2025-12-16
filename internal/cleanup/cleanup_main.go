// File : internal/cleanup/cleanup_main.go
// Deskripsi : Service utama untuk cleanup operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-16

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
