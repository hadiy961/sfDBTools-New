// File : internal/app/profile/executor_adapter.go
// Deskripsi : Adapter untuk memanggil subpackage executor dari Service
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 6 Januari 2026
package profile

import (
	"sfDBTools/internal/app/profile/executor"
	"sfDBTools/internal/domain"
)

func (s *Service) buildExecutor() *executor.Executor {
	e := &executor.Executor{
		Log:                         s.Log,
		ConfigDir:                   s.Config.ConfigDir.DatabaseProfile,
		ProfileInfo:                 s.ProfileInfo,
		ProfileCreate:               s.ProfileCreate,
		ProfileEdit:                 s.ProfileEdit,
		ProfileShow:                 s.ProfileShow,
		ProfileDelete:               s.ProfileDelete,
		OriginalProfileName:         s.OriginalProfileName,
		OriginalProfileInfo:         s.OriginalProfileInfo,
		ValidateProfileInfo:         ValidateProfileInfo,
		FormatConfigToINI:           s.formatConfigToINI,
		ResolveProfileEncryptionKey: resolveProfileEncryptionKey,
	}

	e.RunWizard = func(mode string) error {
		s.ProfileInfo = e.ProfileInfo
		s.ProfileEdit = e.ProfileEdit
		s.ProfileShow = e.ProfileShow
		s.ProfileDelete = e.ProfileDelete
		s.OriginalProfileName = e.OriginalProfileName
		s.OriginalProfileInfo = e.OriginalProfileInfo
		err := s.runWizard(mode)
		// Sinkronkan balik state dari service ke executor,
		// agar step berikutnya (validasi/save) tidak memakai pointer lama.
		e.ProfileInfo = s.ProfileInfo
		e.ProfileEdit = s.ProfileEdit
		e.ProfileShow = s.ProfileShow
		e.ProfileDelete = s.ProfileDelete
		e.OriginalProfileName = s.OriginalProfileName
		e.OriginalProfileInfo = s.OriginalProfileInfo
		return err
	}

	e.DisplayProfileDetails = func() {
		s.ProfileInfo = e.ProfileInfo
		s.ProfileEdit = e.ProfileEdit
		s.ProfileShow = e.ProfileShow
		s.ProfileDelete = e.ProfileDelete
		s.OriginalProfileName = e.OriginalProfileName
		s.OriginalProfileInfo = e.OriginalProfileInfo
		s.DisplayProfileDetails()
	}

	e.CheckConfigurationNameUnique = func(mode string) error {
		s.ProfileInfo = e.ProfileInfo
		s.OriginalProfileName = e.OriginalProfileName
		s.OriginalProfileInfo = e.OriginalProfileInfo
		return s.CheckConfigurationNameUnique(mode)
	}

	e.PromptSelectExistingConfig = func() error {
		err := s.promptSelectExistingConfig()
		e.ProfileInfo = s.ProfileInfo
		e.OriginalProfileName = s.OriginalProfileName
		e.OriginalProfileInfo = s.OriginalProfileInfo
		return err
	}

	e.LoadSnapshotFromPath = func(absPath string) (*domain.ProfileInfo, error) {
		err := s.loadSnapshotFromPath(absPath)
		e.OriginalProfileName = s.OriginalProfileName
		e.OriginalProfileInfo = s.OriginalProfileInfo
		return e.OriginalProfileInfo, err
	}

	return e
}

func (s *Service) applyExecutorState(e *executor.Executor) {
	s.ProfileInfo = e.ProfileInfo
	s.ProfileCreate = e.ProfileCreate
	s.ProfileEdit = e.ProfileEdit
	s.ProfileShow = e.ProfileShow
	s.ProfileDelete = e.ProfileDelete
	s.OriginalProfileName = e.OriginalProfileName
	s.OriginalProfileInfo = e.OriginalProfileInfo
}

func (s *Service) CreateProfile() error {
	e := s.buildExecutor()
	err := e.CreateProfile()
	s.applyExecutorState(e)
	return err
}

func (s *Service) EditProfile() error {
	e := s.buildExecutor()
	err := e.EditProfile()
	s.applyExecutorState(e)
	return err
}

func (s *Service) ShowProfile() error {
	e := s.buildExecutor()
	err := e.ShowProfile()
	s.applyExecutorState(e)
	return err
}

func (s *Service) PromptDeleteProfile() error {
	e := s.buildExecutor()
	err := e.PromptDeleteProfile()
	s.applyExecutorState(e)
	return err
}

func (s *Service) SaveProfile(mode string) error {
	e := s.buildExecutor()
	err := e.SaveProfile(mode)
	s.applyExecutorState(e)
	return err
}
