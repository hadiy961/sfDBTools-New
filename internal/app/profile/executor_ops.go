// File : internal/app/profile/executor_ops.go
// Deskripsi : Adapter ops untuk Executor agar Service tetap tipis (thin orchestrator)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package profile

import (
	"sfdbtools/internal/app/profile/executor"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/app/profile/wizard"
	"sfdbtools/internal/domain"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
)

// executorOps mengenkapsulasi operasi yang dibutuhkan oleh package executor.
// Tujuan: mengurangi jumlah method di Service (thin orchestrator) tanpa mengubah behavior.
//
// Catatan: Ini sengaja membungkus state pointer agar executor/wizard tetap jadi single source of truth.
type executorOps struct {
	Config *appconfig.Config
	Log    applog.Logger
	State  *profilemodel.ProfileState
}

func newExecutorOps(cfg *appconfig.Config, log applog.Logger, state *profilemodel.ProfileState) *executorOps {
	return &executorOps{Config: cfg, Log: log, State: state}
}

func (s *executorOps) configDir() string {
	if s == nil || s.Config == nil {
		return ""
	}
	return s.Config.ConfigDir.DatabaseProfile
}

// =============================================================================
// Implementasi executor.ProfileOps (adapter)
// =============================================================================

func (s *executorOps) NewWizard() *wizard.Runner {
	return wizard.New(
		s.Log,
		s.configDir(),
		s.State,
		s,
		s,
	)
}

// Implementasi dependency wizard (consumer-side interface)

func (s *executorOps) CheckNameUnique(mode string) error {
	return s.CheckConfigurationNameUnique(mode)
}

func (s *executorOps) LoadSnapshot(absPath string) (*domain.ProfileInfo, error) {
	return s.LoadSnapshotFromPath(absPath)
}

func (s *executorOps) SaveProfile(mode string) error {
	e := executor.New(s.Log, s.configDir(), s.State, s)
	return e.SaveProfile(mode)
}

func (s *executorOps) CheckConfigurationNameUnique(mode string) error {
	return s.checkConfigurationNameUnique(mode)
}

func (s *executorOps) LoadSnapshotFromPath(absPath string) (*domain.ProfileInfo, error) {
	err := s.loadSnapshotFromPath(absPath)
	return s.State.OriginalProfileInfo, err
}

func (s *executorOps) PromptSelectExistingConfig() error {
	return s.promptSelectExistingConfig()
}

func (s *executorOps) FormatConfigToINI() string {
	return s.formatConfigToINI()
}
