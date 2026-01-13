// File : internal/profile/wizard/runner.go
// Deskripsi : Runner wizard interaktif untuk create/edit profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 14 Januari 2026
package wizard

import (
	profilehelpers "sfdbtools/internal/app/profile/helpers"
	"fmt"

	profilemodel "sfdbtools/internal/app/profile/model"
	profiledisplay "sfdbtools/internal/app/profile/display"
	"sfdbtools/internal/domain"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
)

// Function types for dependencies
type CheckNameUniqueFn func(mode string) error
type LoadSnapshotFn func(absPath string) (*domain.ProfileInfo, error)

type Runner struct {
	Log       applog.Logger
	ConfigDir string
	State     *profilemodel.ProfileState // Shared state pointer, tidak perlu sync

	// Minimal function dependencies
	CheckNameUnique CheckNameUniqueFn
	LoadSnapshot    LoadSnapshotFn
}

// New creates a new Runner instance
func New(log applog.Logger, configDir string, state *profilemodel.ProfileState, checkName CheckNameUniqueFn, loadSnap LoadSnapshotFn) *Runner {
	return &Runner{
		Log:             log,
		ConfigDir:       configDir,
		State:           state,
		CheckNameUnique: checkName,
		LoadSnapshot:    loadSnap,
	}
}

func (r *Runner) Run(mode string) error {
	if r.Log != nil {
		r.Log.Info(consts.ProfileWizardLogStarted)
	}

	for {
		if mode == consts.ProfileModeEdit {
			if err := r.runEditFlow(); err != nil {
				return err
			}
		} else {
			if err := r.runCreateFlow(mode); err != nil {
				return err
			}
		}

		if r.State.ProfileInfo == nil || r.State.ProfileInfo.Name == "" {
			return fmt.Errorf(consts.ProfileErrProfileNameEmpty)
		}

		// Review
		profiledisplay.DisplayProfileDetails(r.ConfigDir, r.State)

		confirmSave, err := prompt.Confirm(consts.ProfilePromptConfirmSave, true)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if confirmSave {
			break
		}

		confirmRetry, err := prompt.Confirm(consts.ProfilePromptConfirmRetry, false)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if confirmRetry {
			print.PrintWarning(consts.ProfileWizardMsgRestart)
			continue
		}
		return validation.ErrUserCancelled
	}

	print.PrintSuccess(consts.ProfileWizardMsgConfirmAccepted)

	key, source, err := profilehelpers.ResolveProfileEncryptionKey(r.State.ProfileInfo.EncryptionKey, true)
	if err != nil {
		return fmt.Errorf(consts.ProfileErrGetEncryptionPasswordFailedFmt, err)
	}
	if r.Log != nil {
		r.Log.WithField("Sumber Kunci", source).Debug(consts.ProfileLogEncryptionKeyObtained)
	}
	r.State.ProfileInfo.EncryptionKey = key
	return nil
}
