// File : internal/profile/executor/executor.go
// Deskripsi : Executor untuk operasi profile (create/edit/show/delete/save)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026
package executor

import (
	"sfDBTools/internal/services/log"
	"sfDBTools/internal/types"
)

type Executor struct {
	Log       applog.Logger
	ConfigDir string

	ProfileInfo *types.ProfileInfo

	ProfileCreate *types.ProfileCreateOptions
	ProfileEdit   *types.ProfileEditOptions
	ProfileShow   *types.ProfileShowOptions
	ProfileDelete *types.ProfileDeleteOptions

	OriginalProfileName string
	OriginalProfileInfo *types.ProfileInfo

	RunWizard                    func(mode string) error
	DisplayProfileDetails        func()
	CheckConfigurationNameUnique func(mode string) error
	ValidateProfileInfo          func(p *types.ProfileInfo) error
	PromptSelectExistingConfig   func() error
	LoadSnapshotFromPath         func(absPath string) (*types.ProfileInfo, error)

	FormatConfigToINI           func() string
	ResolveProfileEncryptionKey func(existing string, allowPrompt bool) (key string, source string, err error)
}

func (e *Executor) isInteractiveMode() bool {
	if e.ProfileCreate != nil {
		return e.ProfileCreate.Interactive
	}
	if e.ProfileEdit != nil {
		return e.ProfileEdit.Interactive
	}
	if e.ProfileShow != nil {
		return e.ProfileShow.Interactive
	}
	if e.ProfileDelete != nil {
		return e.ProfileDelete.Interactive
	}
	return false
}
