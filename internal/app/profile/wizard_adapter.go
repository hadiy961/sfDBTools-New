// File : internal/app/profile/wizard_adapter.go
// Deskripsi : Adapter untuk memanggil subpackage wizard dari Service
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026
package profile

import (
	"sfDBTools/internal/app/profile/wizard"
	"sfDBTools/internal/domain"
)

func (s *Service) runWizard(mode string) error {
	w := &wizard.Runner{
		Log:                         s.Log,
		ConfigDir:                   s.Config.ConfigDir.DatabaseProfile,
		ProfileInfo:                 s.ProfileInfo,
		ProfileEdit:                 s.ProfileEdit,
		ProfileShow:                 s.ProfileShow,
		OriginalProfileName:         s.OriginalProfileName,
		OriginalProfileInfo:         s.OriginalProfileInfo,
		ResolveProfileEncryptionKey: resolveProfileEncryptionKey,
	}

	// Wiring callback untuk display, agar wizard tetap bisa preview tanpa impor package profile.
	w.DisplayProfileDetails = func() {
		// Sinkronkan state wizard ke service sebelum display.
		s.ProfileInfo = w.ProfileInfo
		s.ProfileEdit = w.ProfileEdit
		s.ProfileShow = w.ProfileShow
		s.OriginalProfileName = w.OriginalProfileName
		s.OriginalProfileInfo = w.OriginalProfileInfo
		s.DisplayProfileDetails()
	}

	w.CheckConfigurationNameUnique = func(mode string) error {
		s.ProfileInfo = w.ProfileInfo
		s.OriginalProfileName = w.OriginalProfileName
		s.OriginalProfileInfo = w.OriginalProfileInfo
		return s.CheckConfigurationNameUnique(mode)
	}

	w.LoadSnapshotFromPath = func(absPath string) (*domain.ProfileInfo, error) {
		s.ProfileInfo = w.ProfileInfo
		err := s.loadSnapshotFromPath(absPath)
		w.OriginalProfileInfo = s.OriginalProfileInfo
		w.OriginalProfileName = s.OriginalProfileName
		return w.OriginalProfileInfo, err
	}

	if err := w.Run(mode); err != nil {
		return err
	}

	// Final sync
	s.ProfileInfo = w.ProfileInfo
	s.ProfileEdit = w.ProfileEdit
	s.ProfileShow = w.ProfileShow
	s.OriginalProfileName = w.OriginalProfileName
	s.OriginalProfileInfo = w.OriginalProfileInfo
	return nil
}
