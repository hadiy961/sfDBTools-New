// File : internal/app/profile/wizard_adapter.go
// Deskripsi : Adapter untuk memanggil subpackage wizard dari Service
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 9 Januari 2026
package profile

import (
	"sfdbtools/internal/app/profile/wizard"
	"sfdbtools/internal/domain"
)

func (s *Service) newWizardRunner() *wizard.Runner {
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

	return w
}

func (s *Service) runWizard(mode string) error {
	w := s.newWizardRunner()

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

// promptCreateRetrySelectedFields dipanggil saat save profile (create) gagal karena koneksi DB invalid
// dan user memilih untuk mengulang input. UX-nya mengikuti profile edit: user memilih field mana yang ingin diubah.
func (s *Service) promptCreateRetrySelectedFields() error {
	w := s.newWizardRunner()

	if err := w.PromptCreateRetrySelectedFields(); err != nil {
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
