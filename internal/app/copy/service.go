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
	ticket string
}

// CopyDatabaseOptions mengkonfigurasi operasi copy database.
type CopyDatabaseOptions struct {
	Profile        *domain.ProfileInfo
	SourceDB       string
	TargetDB       string
	SchemaOnly     bool
	UseConcurrent  bool
	Workers        int
	LimitSpeed     int64
	Force          bool
	BackupFirst    bool
	IncludeGrants  bool
	Verify         bool
	SkipRoutines   bool
	SkipEvents     bool
	SkipTriggers   bool
	NonInteractive bool
}

// CopyDatabasesOptions mengkonfigurasi operasi copy banyak database.
type CopyDatabasesOptions struct {
	CopyDatabaseOptions
	SourceDBs        []string
	Suffix           string
	TargetDBIfSingle string
}

// CopyTableOptions mengkonfigurasi operasi copy tabel tunggal.
type CopyTableOptions struct {
	Profile        *domain.ProfileInfo
	SourceDB       string
	SourceTable    string
	TargetDB       string
	TargetTable    string
	SchemaOnly     bool
	Force          bool
	BackupFirst    bool
	IncludeGrants  bool
	Verify         bool
	NonInteractive bool
}

// CopyTablesConcurrentOptions mengkonfigurasi operasi copy banyak tabel secara paralel.
type CopyTablesConcurrentOptions struct {
	CopyTableOptions
	SourceTables        []string
	Suffix              string
	TargetTableIfSingle string
	Workers             int
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

// SetTicket mengatur ticket number untuk audit.
func (s *Service) SetTicket(ticket string) {
	s.ticket = ticket
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
