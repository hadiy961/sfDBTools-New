// File : internal/profile/executor/executor.go
// Deskripsi : Executor untuk operasi profile (create/edit/show/delete/save)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 14 Januari 2026
package executor

import (
	"fmt"
	"os"
	"strings"

	profilehelpers "sfdbtools/internal/app/profile/helpers"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/app/profile/wizard"
	"sfdbtools/internal/domain"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/runtimecfg"

	"github.com/mattn/go-isatty"
)

// ProfileOps defines minimal operations needed by executor

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
	Log       applog.Logger
	ConfigDir string
	State     *profilemodel.ProfileState // Shared state pointer
	Ops       ProfileOps                 // Operations interface
}

// New creates a new Executor instance
func New(log applog.Logger, configDir string, state *profilemodel.ProfileState, ops ProfileOps) *Executor {
	return &Executor{
		Log:       log,
		ConfigDir: configDir,
		State:     state,
		Ops:       ops,
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

	if e.State.ProfileCreate != nil {
		return e.State.ProfileCreate.Interactive
	}
	if e.State.ProfileEdit != nil {
		return e.State.ProfileEdit.Interactive
	}
	if e.State.ProfileShow != nil {
		return e.State.ProfileShow.Interactive
	}
	if e.State.ProfileDelete != nil {
		return e.State.ProfileDelete.Interactive
	}
	return false
}

func (e *Executor) resolveProfilePath(spec string) (absPath string, name string, err error) {
	if strings.TrimSpace(e.ConfigDir) != "" {
		absPath, name, err = profilehelpers.ResolveConfigPathInDir(e.ConfigDir, spec)
	} else {
		absPath, name, err = profilehelpers.ResolveConfigPath(spec)
	}
	if err != nil {
		return "", "", fmt.Errorf("gagal memproses path konfigurasi: %w", err)
	}
	return absPath, name, nil
}
