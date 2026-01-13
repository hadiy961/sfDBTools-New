// File : internal/profile/service.go
// Deskripsi : Service utama implementation untuk profile operations
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 14 Januari 2026
package profile

import (
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/app/profile/shared"
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

func NewProfileService(cfg *appconfig.Config, logs applog.Logger, profile interface{}) *Service {
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
			svc.State.ProfileCreate = v
			setProfileRefs(&v.ProfileInfo)
		case *profilemodel.ProfileShowOptions:
			svc.State.ProfileShow = v
			setProfileRefs(&v.ProfileInfo)
		case *profilemodel.ProfileEditOptions:
			svc.State.ProfileEdit = v
			setProfileRefs(&v.ProfileInfo)
			// If user provided a file/path via flags, store it as OriginalProfileName
			if v.ProfileInfo.Path != "" {
				svc.State.OriginalProfileName = v.ProfileInfo.Path
			}
		case *profilemodel.ProfileDeleteOptions:
			svc.State.ProfileDelete = v
			setProfileRefs(&v.ProfileInfo)
		default:
			logs.Warn(consts.ProfileLogUnknownProfileTypeInService)
			svc.State.ProfileInfo = &domain.ProfileInfo{}
		}
	} else {
		logs.Warn(consts.ProfileLogUnknownProfileTypeInService)
		svc.State.ProfileInfo = &domain.ProfileInfo{}
	}

	return svc
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
	default:
		return shared.ErrInvalidProfileMode
	}
}
