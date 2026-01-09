// File : internal/profile/executor/executor.go
// Deskripsi : Executor untuk operasi profile (create/edit/show/delete/save)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 6 Januari 2026
package executor

import (
	"os"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/domain"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/runtimecfg"

	"github.com/mattn/go-isatty"
)

type Executor struct {
	Log       applog.Logger
	ConfigDir string

	ProfileInfo *domain.ProfileInfo

	ProfileCreate *profilemodel.ProfileCreateOptions
	ProfileEdit   *profilemodel.ProfileEditOptions
	ProfileShow   *profilemodel.ProfileShowOptions
	ProfileDelete *profilemodel.ProfileDeleteOptions

	OriginalProfileName string
	OriginalProfileInfo *domain.ProfileInfo

	RunWizard                    func(mode string) error
	DisplayProfileDetails        func()
	CheckConfigurationNameUnique func(mode string) error
	ValidateProfileInfo          func(p *domain.ProfileInfo) error
	PromptSelectExistingConfig   func() error
	LoadSnapshotFromPath         func(absPath string) (*domain.ProfileInfo, error)

	FormatConfigToINI           func() string
	ResolveProfileEncryptionKey func(existing string, allowPrompt bool) (key string, source string, err error)
}

func (e *Executor) isInteractiveMode() bool {
	// Hard stop: non-interaktif jika quiet/daemon atau tidak berjalan di TTY.
	if runtimecfg.IsQuiet() || runtimecfg.IsDaemon() {
		return false
	}
	if !isatty.IsTerminal(os.Stdin.Fd()) || !isatty.IsTerminal(os.Stdout.Fd()) {
		return false
	}

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
