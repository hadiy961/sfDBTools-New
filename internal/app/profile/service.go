// File : internal/profile/service.go
// Deskripsi : Service utama implementation untuk profile operations
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 6 Januari 2026
package profile

import (
	"errors"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/domain"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
)

// Error definitions
var (
	ErrInvalidProfileMode = errors.New("mode profile tidak valid")
)

type Service struct {
	Config        *appconfig.Config
	Log           applog.Logger
	ProfileCreate *profilemodel.ProfileCreateOptions
	ProfileInfo   *domain.ProfileInfo
	ProfileShow   *profilemodel.ProfileShowOptions
	ProfileDelete *profilemodel.ProfileDeleteOptions
	ProfileEdit   *profilemodel.ProfileEditOptions
	DBInfo        *domain.DBInfo

	// OriginalProfileName menyimpan nama file profil yang dibuka untuk mode edit.
	OriginalProfileName string
	// OriginalProfileInfo menyimpan salinan data profil sebelum diedit (jika tersedia)
	OriginalProfileInfo *domain.ProfileInfo
}

func NewProfileService(cfg *appconfig.Config, logs applog.Logger, profile interface{}) *Service {
	svc := &Service{
		Log:    logs,
		Config: cfg,
	}

	setProfileRefs := func(info *domain.ProfileInfo) {
		svc.ProfileInfo = info
		svc.DBInfo = &info.DBInfo
	}

	if profile != nil {
		switch v := profile.(type) {
		case *profilemodel.ProfileCreateOptions:
			svc.ProfileCreate = v
			setProfileRefs(&v.ProfileInfo)
		case *profilemodel.ProfileShowOptions:
			svc.ProfileShow = v
			setProfileRefs(&v.ProfileInfo)
		case *profilemodel.ProfileEditOptions:
			svc.ProfileEdit = v
			setProfileRefs(&v.ProfileInfo)
			// If user provided a file/path via flags, store it as OriginalProfileName
			if v.ProfileInfo.Path != "" {
				svc.OriginalProfileName = v.ProfileInfo.Path
			}
		case *profilemodel.ProfileDeleteOptions:
			svc.ProfileDelete = v
			setProfileRefs(&v.ProfileInfo)
		default:
			logs.Warn(consts.ProfileLogUnknownProfileTypeInService)
			svc.ProfileInfo = &domain.ProfileInfo{}
		}
	} else {
		logs.Warn(consts.ProfileLogUnknownProfileTypeInService)
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
		return ErrInvalidProfileMode
	}
}
