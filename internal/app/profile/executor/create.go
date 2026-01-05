// File : internal/profile/executor/create.go
// Deskripsi : Eksekusi pembuatan profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 4 Januari 2026

package executor

import (
	"fmt"

	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/runtimecfg"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

func (e *Executor) CreateProfile() error {
	ui.Headers(consts.ProfileUIHeaderCreate)
	if e.Log != nil {
		e.Log.Info(consts.ProfileLogCreateStarted)
	}

	for {
		if e.ProfileCreate != nil && e.ProfileCreate.Interactive {
			if e.Log != nil {
				e.Log.Info(consts.ProfileLogModeInteractiveEnabled)
			}
			if e.RunWizard == nil {
				return fmt.Errorf(consts.ProfileErrWizardRunnerUnavailable)
			}
			if err := e.RunWizard(consts.ProfileModeCreate); err != nil {
				return err
			}
		} else {
			if e.Log != nil {
				e.Log.Info(consts.ProfileLogModeNonInteractiveEnabled)
				e.Log.Info(consts.ProfileLogValidatingParams)
			}
			if e.ValidateProfileInfo != nil {
				if err := e.ValidateProfileInfo(e.ProfileInfo); err != nil {
					if e.Log != nil {
						e.Log.Errorf(consts.ProfileLogValidationFailedFmt, err)
					}
					return err
				}
			}
			if e.Log != nil {
				e.Log.Info(consts.ProfileLogValidationSuccess)
			}
			if !(runtimecfg.IsQuiet() || runtimecfg.IsDaemon()) {
				if e.DisplayProfileDetails != nil {
					e.DisplayProfileDetails()
				}
			}
		}

		if e.CheckConfigurationNameUnique != nil {
			if err := e.CheckConfigurationNameUnique(consts.ProfileModeCreate); err != nil {
				ui.PrintError(err.Error())
				return err
			}
		}

		if err := e.SaveProfile(consts.ProfileSaveModeCreate); err != nil {
			if err == validation.ErrConnectionFailedRetry {
				retry, err2 := e.handleConnectionFailedRetry(consts.ProfileMsgRetryCreate, consts.ProfileMsgCreateCancelled)
				if err2 != nil {
					return err2
				}
				if retry {
					continue
				}
				return validation.ErrUserCancelled
			}
			return err
		}
		break
	}

	if e.Log != nil {
		e.Log.Info(consts.ProfileLogCreateSuccess)
	}
	return nil
}
