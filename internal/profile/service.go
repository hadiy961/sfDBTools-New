// File : internal/profile/service.go
// Deskripsi : Service utama implementation untuk profile operations
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 5 Januari 2026
package profile

import (
	"errors"
	"sfDBTools/internal/services/config"
	"sfDBTools/internal/services/log"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
)

// Error definitions
var (
	ErrInvalidProfileMode = errors.New("mode profile tidak valid")
)

type Service struct {
	Config        *appconfig.Config
	Log           applog.Logger
	ProfileCreate *types.ProfileCreateOptions
	ProfileInfo   *types.ProfileInfo
	ProfileShow   *types.ProfileShowOptions
	ProfileDelete *types.ProfileDeleteOptions
	ProfileEdit   *types.ProfileEditOptions
	DBInfo        *types.DBInfo

	// OriginalProfileName menyimpan nama file profil yang dibuka untuk mode edit.
	OriginalProfileName string
	// OriginalProfileInfo menyimpan salinan data profil sebelum diedit (jika tersedia)
	OriginalProfileInfo *types.ProfileInfo
}

func NewProfileService(cfg *appconfig.Config, logs applog.Logger, profile interface{}) *Service {
	svc := &Service{
		Log:    logs,
		Config: cfg,
	}

	setProfileRefs := func(info *types.ProfileInfo) {
		svc.ProfileInfo = info
		svc.DBInfo = &info.DBInfo
	}

	if profile != nil {
		switch v := profile.(type) {
		case *types.ProfileCreateOptions:
			svc.ProfileCreate = v
			setProfileRefs(&v.ProfileInfo)
		case *types.ProfileShowOptions:
			svc.ProfileShow = v
			setProfileRefs(&v.ProfileInfo)
		case *types.ProfileEditOptions:
			svc.ProfileEdit = v
			setProfileRefs(&v.ProfileInfo)
			// If user provided a file/path via flags, store it as OriginalProfileName
			if v.ProfileInfo.Path != "" {
				svc.OriginalProfileName = v.ProfileInfo.Path
			}
		case *types.ProfileDeleteOptions:
			svc.ProfileDelete = v
			setProfileRefs(&v.ProfileInfo)
		default:
			logs.Warn(consts.ProfileLogUnknownProfileTypeInService)
			svc.ProfileInfo = &types.ProfileInfo{}
		}
	} else {
		logs.Warn(consts.ProfileLogUnknownProfileTypeInService)
	}

	return svc
}

// ExecuteProfileCommand adalah entry point utama untuk profile execution
func (s *Service) ExecuteProfileCommand(config types.ProfileEntryConfig) error {
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

// isInteractiveMode menentukan apakah service sedang berjalan dalam mode interaktif.
func (s *Service) isInteractiveMode() bool {
	if s.ProfileCreate != nil {
		return s.ProfileCreate.Interactive
	}
	if s.ProfileEdit != nil {
		return s.ProfileEdit.Interactive
	}
	if s.ProfileShow != nil {
		return s.ProfileShow.Interactive
	}
	if s.ProfileDelete != nil {
		return s.ProfileDelete.Interactive
	}
	return false
}
