// File : internal/profile/wizard/runner.go
// Deskripsi : Runner wizard interaktif untuk create/edit profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026
package wizard

import (
	"fmt"

	"sfDBTools/internal/services/log"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

type Runner struct {
	Log       applog.Logger
	ConfigDir string

	ProfileInfo         *types.ProfileInfo
	ProfileEdit         *types.ProfileEditOptions
	ProfileShow         *types.ProfileShowOptions
	OriginalProfileName string
	OriginalProfileInfo *types.ProfileInfo

	DisplayProfileDetails        func()
	CheckConfigurationNameUnique func(mode string) error
	LoadSnapshotFromPath         func(absPath string) (*types.ProfileInfo, error)
	ResolveProfileEncryptionKey  func(existing string, allowPrompt bool) (key string, source string, err error)
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

		if r.ProfileInfo == nil || r.ProfileInfo.Name == "" {
			return fmt.Errorf(consts.ProfileErrProfileNameEmpty)
		}

		// Review
		if r.DisplayProfileDetails != nil {
			r.DisplayProfileDetails()
		}

		confirmSave, err := input.AskYesNo(consts.ProfilePromptConfirmSave, true)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if confirmSave {
			break
		}

		confirmRetry, err := input.AskYesNo(consts.ProfilePromptConfirmRetry, false)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if confirmRetry {
			ui.PrintWarning(consts.ProfileWizardMsgRestart)
			continue
		}
		return validation.ErrUserCancelled
	}

	ui.PrintSuccess(consts.ProfileWizardMsgConfirmAccepted)

	if r.ResolveProfileEncryptionKey == nil {
		return fmt.Errorf(consts.ProfileErrResolveEncryptionKeyUnavailable)
	}

	key, source, err := r.ResolveProfileEncryptionKey(r.ProfileInfo.EncryptionKey, true)
	if err != nil {
		return fmt.Errorf(consts.ProfileErrGetEncryptionPasswordFailedFmt, err)
	}
	if r.Log != nil {
		r.Log.WithField("Sumber Kunci", source).Debug(consts.ProfileLogEncryptionKeyObtained)
	}
	r.ProfileInfo.EncryptionKey = key
	return nil
}
