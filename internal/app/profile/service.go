// File : internal/profile/service.go
// Deskripsi : Service utama implementation untuk profile operations
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 15 Januari 2026
package profile

import (
	profileerrors "sfdbtools/internal/app/profile/errors"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/domain"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
)

type Service struct {
	Config *appconfig.Config
	Log    applog.Logger
	State  *profilemodel.ProfileState // Single source of truth untuk semua shared state
	DBInfo *domain.DBInfo
}

func NewProfileService(cfg *appconfig.Config, logs applog.Logger, profile interface{}) (*Service, error) {
	// Validasi config
	if cfg == nil {
		return nil, profileerrors.ErrConfigIsNil
	}
	if cfg.ConfigDir.DatabaseProfile == "" {
		return nil, profileerrors.ErrConfigDirDatabaseProfileEmpty
	}

	// Fallback untuk logger jika nil
	if logs == nil {
		logs = applog.NullLogger()
	}

	svc := &Service{
		Log:    logs,
		Config: cfg,
		State:  &profilemodel.ProfileState{}, // Initialize shared state
	}

	setProfileRefs := func(info *domain.ProfileInfo) {
		svc.State.ProfileInfo = info
		svc.DBInfo = &info.DBInfo
	}

	if profile != nil {
		switch v := profile.(type) {
		case *profilemodel.ProfileCreateOptions:
			svc.State.Options = v
			setProfileRefs(&v.ProfileInfo)
		case *profilemodel.ProfileShowOptions:
			svc.State.Options = v
			setProfileRefs(&v.ProfileInfo)
		case *profilemodel.ProfileEditOptions:
			svc.State.Options = v
			setProfileRefs(&v.ProfileInfo)
			// If user provided a file/path via flags, store it as OriginalProfileName
			if v.ProfileInfo.Path != "" {
				svc.State.OriginalProfileName = v.ProfileInfo.Path
			}
		case *profilemodel.ProfileDeleteOptions:
			svc.State.Options = v
			setProfileRefs(&v.ProfileInfo)
		case *profilemodel.ProfileCloneOptions:
			svc.State.Options = v
			// For clone, we'll initialize ProfileInfo after loading source
			svc.State.ProfileInfo = &domain.ProfileInfo{}
			svc.DBInfo = &svc.State.ProfileInfo.DBInfo
		default:
			logs.Warn(consts.ProfileLogUnknownProfileTypeInService)
			svc.State.Options = nil
			svc.State.ProfileInfo = &domain.ProfileInfo{}
		}
	} else {
		logs.Warn(consts.ProfileLogUnknownProfileTypeInService)
		svc.State.Options = nil
		svc.State.ProfileInfo = &domain.ProfileInfo{}
	}

	return svc, nil
}

// ExecuteProfileCommand adalah entry point utama untuk profile execution
func (s *Service) ExecuteProfileCommand(config profilemodel.ProfileEntryConfig) error {
	// Log prefix untuk tracking
	if config.LogPrefix != "" {
		s.Log.Infof(consts.ProfileLogStartOperationWithPrefixFmt, config.LogPrefix, config.Mode)
	}

	// Jalankan profile operation berdasarkan mode
	switch config.Mode {
	case consts.ProfileModeCreate:
		return s.CreateProfile()
	case consts.ProfileModeShow:
		return s.ShowProfile()
	case consts.ProfileModeEdit:
		return s.EditProfile()
	case consts.ProfileModeDelete:
		return s.PromptDeleteProfile()
	case consts.ProfileModeClone:
		return s.CloneProfile()
	default:
		return profileerrors.ErrInvalidProfileMode
	}
}
