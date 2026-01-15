// File : internal/profile/executor/executor.go
// Deskripsi : Executor untuk operasi profile (create/edit/show/delete/save)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 15 Januari 2026
package executor

import (
	"os"

	"sfdbtools/internal/app/profile/helpers/paths"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/app/profile/wizard"
	"sfdbtools/internal/domain"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/runtimecfg"

	"github.com/mattn/go-isatty"
)

// WizardProvider menyediakan runner untuk wizard.
type WizardProvider interface {
	NewWizard() *wizard.Runner
}

// ProfileSaver menyimpan profile state ke file.
type ProfileSaver interface {
	SaveProfile(mode string) error
}

// NameUniquenessChecker memvalidasi nama profile unik.
type NameUniquenessChecker interface {
	CheckConfigurationNameUnique(mode string) error
}

// SnapshotLoader memuat snapshot dari file profile.
type SnapshotLoader interface {
	LoadSnapshotFromPath(absPath string) (*domain.ProfileInfo, error)
}

// ExistingConfigSelector menangani pemilihan profile existing secara interaktif.
type ExistingConfigSelector interface {
	PromptSelectExistingConfig() error
}

// ConfigFormatter mengubah state profile menjadi format INI.
type ConfigFormatter interface {
	FormatConfigToINI() string
}

// ProfileOps adalah composite interface untuk kebutuhan Executor saat ini.
// Catatan: Executor methods idealnya hanya bergantung pada sub-interface yang dibutuhkan.
type ProfileOps interface {
	WizardProvider
	ProfileSaver
	NameUniquenessChecker
	SnapshotLoader
	ExistingConfigSelector
	ConfigFormatter
}

type Executor struct {
	Log          applog.Logger
	ConfigDir    string
	State        *profilemodel.ProfileState // Shared state pointer
	Ops          ProfileOps                 // Operations interface
	pathResolver *paths.PathResolver        // Path resolution helper
}

// New creates a new Executor instance
func New(log applog.Logger, configDir string, state *profilemodel.ProfileState, ops ProfileOps) *Executor {
	if log == nil {
		log = applog.NullLogger()
	}
	return &Executor{
		Log:          log,
		ConfigDir:    configDir,
		State:        state,
		Ops:          ops,
		pathResolver: paths.NewPathResolver(configDir),
	}
}

func (e *Executor) isInteractiveMode() bool {
	// Hard stop: non-interaktif jika quiet/daemon atau tidak berjalan di TTY.
	if runtimecfg.IsQuiet() || runtimecfg.IsDaemon() {
		return false
	}
	if !isatty.IsTerminal(os.Stdin.Fd()) || !isatty.IsTerminal(os.Stdout.Fd()) {
		return false
	}
	return e.State.IsInteractive()
}

func (e *Executor) resolveProfilePath(spec string) (absPath string, name string, err error) {
	return e.pathResolver.Resolve(spec)
}
