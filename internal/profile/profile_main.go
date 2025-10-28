package profile

import (
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
)

type Service struct {
	Config        *appconfig.Config
	ProfileCreate *types.ProfileCreateOptions
	ProfileInfo   *types.ProfileInfo
	ProfileShow   *types.ProfileShowOptions
	ProfileDelete *types.ProfileDeleteOptions
	ProfileEdit   *types.ProfileEditOptions
	DBInfo        *types.DBInfo
	Log           applog.Logger
	// OriginalProfileName menyimpan nama file profil yang dibuka untuk mode edit.
	OriginalProfileName string
	// OriginalProfileInfo menyimpan salinan data profil sebelum diedit (jika tersedia)
	OriginalProfileInfo *types.ProfileInfo
}

func NewService(cfg *appconfig.Config, logs applog.Logger, profile interface{}) *Service {
	svc := &Service{
		Log:    logs,
		Config: cfg, // Perbaikan: set field Config agar tidak nil
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
			// Tambahkan inisialisasi untuk ProfileDeleteOptions jika diperlukan
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
