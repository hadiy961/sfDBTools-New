// File : internal/profile/service.go
// Deskripsi : Service utama implementation untuk profile operations
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 17 Desember 2025

package profile

import (
	"errors"
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
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

	if profile != nil {
		switch v := profile.(type) {
		case *types.ProfileCreateOptions:
			svc.ProfileCreate = v
			svc.ProfileInfo = &v.ProfileInfo
			svc.DBInfo = &v.ProfileInfo.DBInfo
		case *types.ProfileShowOptions:
			svc.ProfileShow = v
			svc.ProfileInfo = &v.ProfileInfo
			svc.DBInfo = &v.ProfileInfo.DBInfo
		case *types.ProfileEditOptions:
			svc.ProfileEdit = v
			svc.ProfileInfo = &v.ProfileInfo
			svc.DBInfo = &v.ProfileInfo.DBInfo
			// If user provided a file/path via flags, store it as OriginalProfileName
			if v.ProfileInfo.Path != "" {
				svc.OriginalProfileName = v.ProfileInfo.Path
			}
		case *types.ProfileDeleteOptions:
			svc.ProfileDelete = v
			svc.ProfileInfo = &v.ProfileInfo
			svc.DBInfo = &v.ProfileInfo.DBInfo
		default:
			logs.Warn("Tipe profil tidak dikenali dalam Service")
			svc.ProfileInfo = &types.ProfileInfo{}
		}
	} else {
		logs.Warn("Tipe profil tidak dikenali dalam Service")
	}

	return svc
}

// ExecuteProfileCommand adalah entry point utama untuk profile execution
func (s *Service) ExecuteProfileCommand(config types.ProfileEntryConfig) error {
	// Log prefix untuk tracking
	if config.LogPrefix != "" {
		s.Log.Infof("[%s] Memulai profile operation dengan mode: %s", config.LogPrefix, config.Mode)
	}

	// Tampilkan options jika diminta
	if config.ShowOptions {
		s.displayProfileOptions()
	}

	// Jalankan profile operation berdasarkan mode
	switch config.Mode {
	case "create":
		return s.CreateProfile()
	case "show":
		return s.ShowProfile()
	case "edit":
		return s.EditProfile()
	case "delete":
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
	// Default ke false untuk mode lain (show, delete)
	return false
}
