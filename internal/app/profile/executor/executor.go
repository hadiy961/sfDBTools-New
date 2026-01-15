// File : internal/profile/executor/executor.go
// Deskripsi : Executor untuk operasi profile (create/edit/show/delete/save)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 15 Januari 2026
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
// ProfileOps adalah interface minimal untuk kebutuhan Executor saat ini.
// Catatan: Sebelumnya dipecah menjadi banyak sub-interface (ISP), namun itu berlebihan
// karena (saat ini) hanya ada satu implementasi ops. Kita pertahankan 1 interface
// untuk mengurangi bloat (YAGNI) tanpa mengubah behavior.
type ProfileOps interface {
	NewWizard() *wizard.Runner
	SaveProfile(mode string) error
	CheckConfigurationNameUnique(mode string) error
	LoadSnapshotFromPath(absPath string) (*domain.ProfileInfo, error)
	PromptSelectExistingConfig() error
	FormatConfigToINI() string
}

type Executor struct {
	Log       applog.Logger
	ConfigDir string
	State     *profilemodel.ProfileState // Shared state pointer
	Ops       ProfileOps                 // Operations interface
}

// New creates a new Executor instance
func New(log applog.Logger, configDir string, state *profilemodel.ProfileState, ops ProfileOps) *Executor {
	if log == nil {
		log = applog.NullLogger()
	}
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
	return e.State.IsInteractive()
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
