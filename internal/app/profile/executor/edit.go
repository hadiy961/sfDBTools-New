// File : internal/app/profile/executor/edit.go
// Deskripsi : Eksekusi edit profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026

package executor

import (
	"fmt"
	"strings"

	"sfDBTools/internal/app/profile/shared"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

func (e *Executor) EditProfile() error {
	ui.Headers(consts.ProfileUIHeaderEdit)

	for {
		if e.ProfileEdit != nil && e.ProfileEdit.Interactive {
			if e.Log != nil {
				e.Log.Info(consts.ProfileLogModeInteractiveEnabled)
			}
			if e.RunWizard == nil {
				return fmt.Errorf(consts.ProfileErrWizardRunnerUnavailable)
			}
			if err := e.RunWizard(consts.ProfileModeEdit); err != nil {
				if err == validation.ErrUserCancelled {
					if e.Log != nil {
						e.Log.Warn(consts.ProfileLogEditCancelledByUser)
					}
					return validation.ErrUserCancelled
				}
				if e.Log != nil {
					e.Log.Warn(fmt.Sprintf(consts.ProfileLogEditFailedFmt, err))
				}
				return err
			}
		} else {
			if e.Log != nil {
				e.Log.Info(consts.ProfileLogModeNonInteractiveShort)
			}
			if strings.TrimSpace(e.OriginalProfileName) == "" {
				return fmt.Errorf(consts.ProfileErrNonInteractiveProfileRequired)
			}

			overrideDB := e.ProfileInfo.DBInfo
			overrideSSH := e.ProfileInfo.SSHTunnel

			absPath, name, err := helper.ResolveConfigPath(e.OriginalProfileName)
			if err != nil {
				return err
			}
			if !fsops.PathExists(absPath) {
				ui.PrintWarning(fmt.Sprintf(consts.ProfileWarnConfigFileNotFoundFmt, absPath))
				return fmt.Errorf(consts.ProfileErrConfigFileNotFoundFmt, absPath)
			}

			e.ProfileInfo.Name = name
			e.ProfileInfo.Path = absPath
			e.OriginalProfileName = name

			if e.Log != nil {
				e.Log.Info(consts.ProfileLogConfigFileFoundTryLoad)
			}
			if e.LoadSnapshotFromPath != nil {
				if snap, err := e.LoadSnapshotFromPath(absPath); err != nil {
					ui.PrintWarning(fmt.Sprintf(consts.ProfileWarnLoadFileContentFailedProceedFmt, absPath, err))
				} else {
					e.OriginalProfileInfo = snap
				}
			}
			if e.Log != nil {
				e.Log.Info(consts.ProfileLogConfigFileLoaded)
			}

			shared.ApplySnapshotAsBaseline(e.ProfileInfo, e.OriginalProfileInfo)
			shared.ApplyDBOverrides(e.ProfileInfo, overrideDB)
			shared.ApplySSHOverrides(e.ProfileInfo, overrideSSH)

			if e.Log != nil {
				e.Log.Info(consts.ProfileLogValidatingParams)
			}
			if e.ProfileEdit != nil && strings.TrimSpace(e.ProfileEdit.NewName) != "" {
				newName := helper.TrimProfileSuffix(strings.TrimSpace(e.ProfileEdit.NewName))
				if newName == "" {
					return fmt.Errorf(consts.ProfileErrNewNameEmpty)
				}
				if strings.Contains(newName, "/") || strings.Contains(newName, "\\") {
					return fmt.Errorf(consts.ProfileErrNewNameContainsPath)
				}
				e.ProfileInfo.Name = newName
			}

			if e.ValidateProfileInfo != nil {
				if err := e.ValidateProfileInfo(e.ProfileInfo); err != nil {
					if e.Log != nil {
						e.Log.Errorf(consts.ProfileLogValidationFailedFmt, err)
					}
					return err
				}
			}
			if e.DisplayProfileDetails != nil {
				e.DisplayProfileDetails()
			}
		}

		if e.CheckConfigurationNameUnique != nil {
			if err := e.CheckConfigurationNameUnique(consts.ProfileModeEdit); err != nil {
				ui.PrintError(err.Error())
				return err
			}
		}

		if err := e.SaveProfile(consts.ProfileSaveModeEdit); err != nil {
			if err == validation.ErrConnectionFailedRetry {
				retry, err2 := e.handleConnectionFailedRetry(consts.ProfileMsgRetryEdit, consts.ProfileMsgEditCancelled)
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
		e.Log.Info(consts.ProfileLogWizardInteractiveFinished)
	}
	return nil
}
