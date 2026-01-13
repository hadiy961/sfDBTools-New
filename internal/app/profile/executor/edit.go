// File : internal/app/profile/executor/edit.go
// Deskripsi : Eksekusi edit profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 6 Januari 2026

package executor

import (
	"fmt"
	"strings"

	"sfdbtools/internal/app/profile/shared"
	profilevalidation "sfdbtools/internal/app/profile/validation"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/fsops"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	profilehelpers "sfdbtools/internal/app/profile/helpers"
)

func (e *Executor) EditProfile() error {
	for {
		if e.State.ProfileEdit != nil && e.State.ProfileEdit.Interactive {
			if e.Log != nil {
				e.Log.Info(consts.ProfileLogModeInteractiveEnabled)
			}
			if err := e.Ops.NewWizard().Run(consts.ProfileModeEdit); err != nil {
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
			if strings.TrimSpace(e.State.OriginalProfileName) == "" {
				return fmt.Errorf(consts.ProfileErrNonInteractiveProfileRequired)
			}

			overrideDB := e.State.ProfileInfo.DBInfo
			overrideSSH := e.State.ProfileInfo.SSHTunnel

			absPath, name, err := profilehelpers.ResolveConfigPath(e.State.OriginalProfileName)
			if err != nil {
				return err
			}
			if !fsops.PathExists(absPath) {
				print.PrintWarning(fmt.Sprintf(consts.ProfileWarnConfigFileNotFoundFmt, absPath))
				return fmt.Errorf(consts.ProfileErrConfigFileNotFoundFmt, absPath)
			}

			e.State.ProfileInfo.Name = name
			e.State.ProfileInfo.Path = absPath
			e.State.OriginalProfileName = name

			if e.Log != nil {
				e.Log.Info(consts.ProfileLogConfigFileFoundTryLoad)
			}
			if e.Ops.LoadSnapshotFromPath != nil {
				if snap, err := e.Ops.LoadSnapshotFromPath(absPath); err != nil {
					print.PrintWarning(fmt.Sprintf(consts.ProfileWarnLoadFileContentFailedProceedFmt, absPath, err))
				} else {
					e.State.OriginalProfileInfo = snap
				}
			}
			if e.Log != nil {
				e.Log.Info(consts.ProfileLogConfigFileLoaded)
			}

			shared.ApplySnapshotAsBaseline(e.State.ProfileInfo, e.State.OriginalProfileInfo)
			shared.ApplyDBOverrides(e.State.ProfileInfo, overrideDB)
			shared.ApplySSHOverrides(e.State.ProfileInfo, overrideSSH)

			if e.Log != nil {
				e.Log.Info(consts.ProfileLogValidatingParams)
			}
			if e.State.ProfileEdit != nil && strings.TrimSpace(e.State.ProfileEdit.NewName) != "" {
				newName := shared.TrimProfileSuffix(strings.TrimSpace(e.State.ProfileEdit.NewName))
				if newName == "" {
					return fmt.Errorf(consts.ProfileErrNewNameEmpty)
				}
				if strings.Contains(newName, "/") || strings.Contains(newName, "\\") {
					return fmt.Errorf(consts.ProfileErrNewNameContainsPath)
				}
				e.State.ProfileInfo.Name = newName
			}

			if profilevalidation.ValidateProfileInfo != nil {
				if err := profilevalidation.ValidateProfileInfo(e.State.ProfileInfo); err != nil {
					if e.Log != nil {
						e.Log.Errorf(consts.ProfileLogValidationFailedFmt, err)
					}
					return err
				}
			}
		}

		if e.Ops.CheckConfigurationNameUnique != nil {
			if err := e.Ops.CheckConfigurationNameUnique(consts.ProfileModeEdit); err != nil {
				print.PrintError(err.Error())
				return err
			}
		}

		// Jika user memilih untuk merubah kunci enkripsi (rotasi), gunakan key baru saat save.
		// Decrypt snapshot sudah dilakukan sebelumnya memakai key lama.
		if e.State.ProfileEdit != nil && strings.TrimSpace(e.State.ProfileEdit.NewProfileKey) != "" {
			e.State.ProfileInfo.EncryptionKey = strings.TrimSpace(e.State.ProfileEdit.NewProfileKey)
			e.State.ProfileInfo.EncryptionSource = strings.TrimSpace(e.State.ProfileEdit.NewProfileKeySource)
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
		if !e.isInteractiveMode() {
			e.Log.Info(consts.ProfileLogWizardInteractiveFinished)
		}
	}
	return nil
}
