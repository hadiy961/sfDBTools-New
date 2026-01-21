// File : internal/app/profile/service_methods.go
// Deskripsi : Service methods for wizard/executor integration (P2 refactored)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package profile

import (
	"sfdbtools/internal/app/profile/executor"
)

// CreateProfile creates a new profile
func (s *Service) CreateProfile() error {
	ops := newExecutorOps(s.Config, s.Log, s.State)
	e := executor.New(s.Log, s.Config, ops.configDir(), s.State, ops)
	return e.CreateProfile()
}

// EditProfile edits an existing profile
func (s *Service) EditProfile() error {
	ops := newExecutorOps(s.Config, s.Log, s.State)
	e := executor.New(s.Log, s.Config, ops.configDir(), s.State, ops)
	return e.EditProfile()
}

// ShowProfile shows profile details
func (s *Service) ShowProfile() error {
	ops := newExecutorOps(s.Config, s.Log, s.State)
	e := executor.New(s.Log, s.Config, ops.configDir(), s.State, ops)
	return e.ShowProfile()
}

// PromptDeleteProfile deletes a profile
func (s *Service) PromptDeleteProfile() error {
	ops := newExecutorOps(s.Config, s.Log, s.State)
	e := executor.New(s.Log, s.Config, ops.configDir(), s.State, ops)
	return e.PromptDeleteProfile()
}

// CloneProfile clones an existing profile
func (s *Service) CloneProfile() error {
	ops := newExecutorOps(s.Config, s.Log, s.State)
	e := executor.New(s.Log, s.Config, ops.configDir(), s.State, ops)
	return e.CloneProfile()
}
