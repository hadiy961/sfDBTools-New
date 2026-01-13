// File : internal/profile/executor/create.go
// Deskripsi : Eksekusi pembuatan profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 9 Januari 2026

package executor

import (
	profiledisplay "sfdbtools/internal/app/profile/display"
	profilevalidation "sfdbtools/internal/app/profile/validation"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
)

func (e *Executor) CreateProfile() error {
	isInteractive := e.isInteractiveMode()
	if e.Log != nil && !isInteractive {
		e.Log.Info(consts.ProfileLogCreateStarted)
	}

	// Jika retry save karena koneksi DB invalid, kita tidak ingin restart wizard dari awal.
	skipWizard := false

	for {
		if !skipWizard && e.State.ProfileCreate != nil && e.State.ProfileCreate.Interactive {
			// Mode interaktif: hindari log Info agar tidak mengganggu prompt.
			if err := e.Ops.NewWizard().Run(consts.ProfileModeCreate); err != nil {
				return err
			}
		} else if !skipWizard {
			if e.Log != nil {
				e.Log.Info(consts.ProfileLogModeNonInteractiveEnabled)
				e.Log.Info(consts.ProfileLogValidatingParams)
			}
			if err := profilevalidation.ValidateProfileInfo(e.State.ProfileInfo); err != nil {
				if e.Log != nil {
					e.Log.Errorf(consts.ProfileLogValidationFailedFmt, err)
				}
				return err
			}
			if e.Log != nil {
				e.Log.Info(consts.ProfileLogValidationSuccess)
			}
			if !(runtimecfg.IsQuiet() || runtimecfg.IsDaemon()) {
				profiledisplay.DisplayProfileDetails(e.ConfigDir, e.State)
			}
		}
		skipWizard = false

		if err := e.Ops.CheckConfigurationNameUnique(consts.ProfileModeCreate); err != nil {
			print.PrintError(err.Error())
			return err
		}

		if err := e.Ops.SaveProfile(consts.ProfileSaveModeCreate); err != nil {
			if err == validation.ErrConnectionFailedRetry {
				retry, err2 := e.handleConnectionFailedRetry(consts.ProfileMsgRetryCreate, consts.ProfileMsgCreateCancelled)
				if err2 != nil {
					return err2
				}
				if retry {
					// UX: setelah retry, tampilkan selector field (mirip profile edit), bukan restart wizard dari awal.
					if e.isInteractiveMode() {
						if err := e.Ops.NewWizard().PromptCreateRetrySelectedFields(); err != nil {
							return err
						}
						skipWizard = true
					}
					continue
				}
				return validation.ErrUserCancelled
			}
			return err
		}
		break
	}
	return nil
}
