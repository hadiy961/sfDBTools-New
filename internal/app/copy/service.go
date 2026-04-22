package copy

import (
	"sfdbtools/internal/app/profile/helpers/loader"
	"sfdbtools/internal/domain"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/errorlog"
)

// Service adalah orchestrator utama untuk fitur copy.
type Service struct {
	log    applog.Logger
	cfg    *appconfig.Config
	errLog *errorlog.ErrorLogger
}

// NewService membuat instance baru dari Service.
func NewService(log applog.Logger, cfg *appconfig.Config) *Service {
	logDir := consts.DefaultLogDir
	if cfg != nil && cfg.Log.Output.File.Dir != "" {
		logDir = cfg.Log.Output.File.Dir
	}
	return &Service{
		log:    log,
		cfg:    cfg,
		errLog: errorlog.NewErrorLogger(log, logDir, consts.FeatureBackup),
	}
}

// LoadProfile me-load profil database dengan dukungan enkripsi (--profile-key).
func (s *Service) LoadProfile(profileName, profileKey string, allowInteractive bool) (*domain.ProfileInfo, error) {
	configDir := ""
	if s.cfg != nil {
		configDir = s.cfg.ConfigDir.DatabaseProfile
	}

	// Jika interaktif diizinkan dan profil kosong, gunakan LoadSourceProfile untuk picker
	if allowInteractive && profileName == "" {
		return loader.LoadSourceProfile(configDir, profileName, profileKey, true)
	}

	return loader.ResolveAndLoadProfile(loader.ProfileLoadOptions{
		ConfigDir:        configDir,
		ProfilePath:      profileName,
		ProfileKey:       profileKey,
		RequireProfile:   true,
		AllowInteractive: allowInteractive,
	})
}
