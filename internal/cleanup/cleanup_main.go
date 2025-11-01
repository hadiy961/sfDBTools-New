package cleanup

import (
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
)

// Service adalah service untuk database scanning
type Service struct {
	Logger         applog.Logger
	Config         *appconfig.Config
	CleanupOptions types.CleanupOptions
}

// NewService membuat instance baru dari Service
func NewCleanupService(logger applog.Logger, config *appconfig.Config) *Service {
	return &Service{
		Logger: logger,
		Config: config,
	}
}

// SetCleanupOptions mengatur opsi cleanup
func (s *Service) SetCleanupOptions(opts types.CleanupOptions) {
	s.CleanupOptions = opts
}
