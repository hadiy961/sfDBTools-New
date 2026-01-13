// File : internal/app/profile/service_methods.go
// Deskripsi : Service methods for wizard/executor integration (P2 refactored)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package profile

import (
	"sfdbtools/internal/app/profile/executor"
	"sfdbtools/internal/app/profile/wizard"
	"sfdbtools/internal/domain"
)

// Service implements executor.ProfileOps interface
var _ executor.ProfileOps = (*Service)(nil)

// CreateProfile creates a new profile
func (s *Service) CreateProfile() error {
	e := executor.New(s.Log, s.Config.ConfigDir.DatabaseProfile, s.State, s)
	return e.CreateProfile()
}

// EditProfile edits an existing profile
func (s *Service) EditProfile() error {
	e := executor.New(s.Log, s.Config.ConfigDir.DatabaseProfile, s.State, s)
	return e.EditProfile()
}

// ShowProfile shows profile details
func (s *Service) ShowProfile() error {
	e := executor.New(s.Log, s.Config.ConfigDir.DatabaseProfile, s.State, s)
	return e.ShowProfile()
}

// PromptDeleteProfile deletes a profile
func (s *Service) PromptDeleteProfile() error {
	e := executor.New(s.Log, s.Config.ConfigDir.DatabaseProfile, s.State, s)
	return e.PromptDeleteProfile()
}

// SaveProfile saves current profile state
func (s *Service) SaveProfile(mode string) error {
	e := executor.New(s.Log, s.Config.ConfigDir.DatabaseProfile, s.State, s)
	return e.SaveProfile(mode)
}

// ProfileOps interface implementations

// NewWizard creates a new wizard runner
func (s *Service) NewWizard() *wizard.Runner {
	return wizard.New(
		s.Log,
		s.Config.ConfigDir.DatabaseProfile,
		s.State,
		s.CheckConfigurationNameUnique,
		s.LoadSnapshotFromPath,
	)
}

// CheckConfigurationNameUnique checks if configuration name is unique
func (s *Service) CheckConfigurationNameUnique(mode string) error {
	return s.checkConfigurationNameUnique(mode)
}

// LoadSnapshotFromPath loads profile snapshot from path
func (s *Service) LoadSnapshotFromPath(absPath string) (*domain.ProfileInfo, error) {
	err := s.loadSnapshotFromPath(absPath)
	return s.State.OriginalProfileInfo, err
}

// PromptSelectExistingConfig prompts user to select existing config
func (s *Service) PromptSelectExistingConfig() error {
	return s.promptSelectExistingConfig()
}

// FormatConfigToINI formats config to INI string
func (s *Service) FormatConfigToINI() string {
	return s.formatConfigToINI()
}
