// File : internal/profile/executor/create.go
// Deskripsi : Eksekusi pembuatan profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 14 Januari 2026

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
	if !isInteractive {
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
			{
				e.Log.Info(consts.ProfileLogModeNonInteractiveEnabled)
				e.Log.Info(consts.ProfileLogValidatingParams)
			}
			if err := profilevalidation.ValidateProfileInfo(e.State.ProfileInfo); err != nil {
				{
					e.Log.Errorf(consts.ProfileLogValidationFailedFmt, err)
				}
				return err
			}
			{
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
			retry, err2 := e.handleConnectionFailedRetryIfNeeded(err, consts.ProfileMsgRetryCreate, consts.ProfileMsgCreateCancelled)
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
			// Defensive: seharusnya tidak pernah sampai sini (cancel return error).
			return validation.ErrUserCancelled
		}
		break
	}
	return nil
}
