// File : internal/app/profile/executor/edit.go
// Deskripsi : Eksekusi edit profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 15 Januari 2026

package executor

import (
	"fmt"
	"strings"

	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/app/profile/merger"
	profilevalidation "sfdbtools/internal/app/profile/validation"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/fsops"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
)

func (e *Executor) EditProfile() error {
	for {
		editOpts, _ := e.State.EditOptions()
		if editOpts != nil && editOpts.Interactive {
			{
				e.Log.Info(consts.ProfileLogModeInteractiveEnabled)
			}
			if e.Ops == nil {
				return fmt.Errorf(consts.ProfileErrWizardRunnerUnavailable)
			}
			runner := e.Ops.NewWizard()
			if runner == nil {
				return fmt.Errorf(consts.ProfileErrWizardRunnerUnavailable)
			}
			if err := runner.Run(consts.ProfileModeEdit); err != nil {
				if err == validation.ErrUserCancelled {
					{
						e.Log.Warn(consts.ProfileLogEditCancelledByUser)
					}
					return validation.ErrUserCancelled
				}
				{
					e.Log.Warn(fmt.Sprintf(consts.ProfileLogEditFailedFmt, err))
				}
				return err
			}
		} else {
			{
				e.Log.Info(consts.ProfileLogModeNonInteractiveShort)
			}
			if strings.TrimSpace(e.State.OriginalProfileName) == "" {
				return fmt.Errorf(consts.ProfileErrNonInteractiveProfileRequired)
			}

			overrideDB := e.State.ProfileInfo.DBInfo
			overrideSSH := e.State.ProfileInfo.SSHTunnel

			absPath, name, err := e.resolveProfilePath(e.State.OriginalProfileName)
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

			{
				e.Log.Info(consts.ProfileLogConfigFileFoundTryLoad)
			}
			if e.Ops != nil {
				if snap, err := e.Ops.LoadSnapshotFromPath(absPath); err != nil {
					print.PrintWarning(fmt.Sprintf(consts.ProfileWarnLoadFileContentFailedProceedFmt, absPath, err))
				} else {
					e.State.OriginalProfileInfo = snap
				}
			}
			{
				e.Log.Info(consts.ProfileLogConfigFileLoaded)
			}

			merger.ApplySnapshotAsBaseline(e.State.ProfileInfo, e.State.OriginalProfileInfo)
			merger.ApplyDBOverrides(e.State.ProfileInfo, overrideDB)
			merger.ApplySSHOverrides(e.State.ProfileInfo, overrideSSH)

			{
				e.Log.Info(consts.ProfileLogValidatingParams)
			}
			if editOpts != nil && strings.TrimSpace(editOpts.NewName) != "" {
				newName := profileconn.TrimProfileSuffix(strings.TrimSpace(editOpts.NewName))
				if newName == "" {
					return fmt.Errorf(consts.ProfileErrNewNameEmpty)
				}
				if strings.Contains(newName, "/") || strings.Contains(newName, "\\") {
					return fmt.Errorf(consts.ProfileErrNewNameContainsPath)
				}
				e.State.ProfileInfo.Name = newName
			}

			if err := profilevalidation.ValidateProfileInfo(e.State.ProfileInfo); err != nil {
				{
					e.Log.Errorf(consts.ProfileLogValidationFailedFmt, err)
				}
				return err
			}
		}

		if e.Ops == nil {
			return fmt.Errorf(consts.ProfileErrWizardRunnerUnavailable)
		}
		if err := e.Ops.CheckConfigurationNameUnique(consts.ProfileModeEdit); err != nil {
			print.PrintError(err.Error())
			return err
		}

		// Jika user memilih untuk merubah kunci enkripsi (rotasi), gunakan key baru saat save.
		// Decrypt snapshot sudah dilakukan sebelumnya memakai key lama.
		if editOpts != nil && strings.TrimSpace(editOpts.NewProfileKey) != "" {
			e.State.ProfileInfo.EncryptionKey = strings.TrimSpace(editOpts.NewProfileKey)
			e.State.ProfileInfo.EncryptionSource = strings.TrimSpace(editOpts.NewProfileKeySource)
		}

		// Jika tidak ada perubahan meaningful, jangan tulis ulang file.
		if e.State != nil && e.State.OriginalProfileInfo != nil && !e.State.HasMeaningfulChanges() {
			return validation.ErrUserCancelled
		}

		if err := e.SaveProfile(consts.ProfileSaveModeEdit); err != nil {
			retry, err2 := e.handleConnectionFailedRetryIfNeeded(err, consts.ProfileMsgRetryEdit, consts.ProfileMsgEditCancelled)
			if err2 != nil {
				return err2
			}
			if retry {
				continue
			}
			// Defensive: seharusnya tidak pernah sampai sini (cancel return error).
			return validation.ErrUserCancelled
		}
		break
	}

	{
		if !e.isInteractiveMode() {
			e.Log.Info(consts.ProfileLogWizardInteractiveFinished)
		}
	}
	return nil
}
