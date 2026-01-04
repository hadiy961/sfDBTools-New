// File : internal/profile/executor/save.go
// Deskripsi : Simpan profile ke file (terenkripsi)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026

package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/profile/shared"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	profilehelper "sfDBTools/pkg/helper/profile"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

func (e *Executor) SaveProfile(mode string) error {
	if e.Log != nil {
		e.Log.Info(consts.ProfileLogStartSave)
	}
	isInteractive := e.isInteractiveMode()

	var baseDir string
	var originalAbsPath string
	if mode == consts.ProfileSaveModeEdit && e.ProfileInfo.Path != "" && filepath.IsAbs(e.ProfileInfo.Path) {
		originalAbsPath = e.ProfileInfo.Path
		baseDir = filepath.Dir(e.ProfileInfo.Path)
	} else {
		baseDir = e.ConfigDir
	}

	if !fsops.DirExists(baseDir) {
		if err := fsops.CreateDirIfNotExist(baseDir); err != nil {
			return fmt.Errorf(consts.ProfileErrCreateConfigDirFailedFmt, err)
		}
	}

	if c, err := profilehelper.ConnectWithProfile(e.ProfileInfo, consts.DefaultInitialDatabase); err != nil {
		if !isInteractive {
			return err
		}
		continueAnyway, askErr := input.AskYesNo(consts.ProfileSavePromptContinueDespiteDBFail, false)
		if askErr != nil {
			return validation.HandleInputError(askErr)
		}
		if !continueAnyway {
			return validation.ErrConnectionFailedRetry
		}
		ui.PrintWarning(consts.ProfileSaveWarnSavingWithInvalidConn)
	} else {
		c.Close()
		if e.Log != nil {
			e.Log.Info(consts.ProfileLogDBConnectionValid)
		}
	}

	if e.FormatConfigToINI == nil {
		return fmt.Errorf(consts.ProfileErrFormatINIUnavailable)
	}
	iniContent := e.FormatConfigToINI()

	if e.ResolveProfileEncryptionKey == nil {
		return fmt.Errorf(consts.ProfileErrResolveEncryptionKeyUnavailable)
	}
	key, _, err := e.ResolveProfileEncryptionKey(e.ProfileInfo.EncryptionKey, isInteractive)
	if err != nil {
		return fmt.Errorf(consts.ProfileErrEncryptionKeyUnavailableFmt, err)
	}
	e.ProfileInfo.EncryptionKey = strings.TrimSpace(key)

	encryptedContent, err := encrypt.EncryptAES([]byte(iniContent), []byte(e.ProfileInfo.EncryptionKey))
	if err != nil {
		return fmt.Errorf(consts.ProfileErrEncryptConfigFailedFmt, err)
	}

	if mode == consts.ProfileSaveModeEdit && e.ProfileEdit != nil && strings.TrimSpace(e.ProfileEdit.NewName) != "" {
		if err := validation.ValidateProfileName(e.ProfileEdit.NewName); err != nil {
			return err
		}
		e.ProfileInfo.Name = e.ProfileEdit.NewName
	}
	if err := validation.ValidateProfileName(e.ProfileInfo.Name); err != nil {
		return err
	}

	e.ProfileInfo.Name = helper.TrimProfileSuffix(e.ProfileInfo.Name)
	newFileName := shared.BuildProfileFileName(e.ProfileInfo.Name)
	newFilePath := filepath.Join(baseDir, newFileName)

	if mode == consts.ProfileSaveModeEdit && e.OriginalProfileName != "" && e.OriginalProfileName != e.ProfileInfo.Name {
		if err := fsops.WriteFile(newFilePath, encryptedContent); err != nil {
			return fmt.Errorf(consts.ProfileErrWriteNewConfigFailedFmt, err)
		}

		oldFilePath := originalAbsPath
		if oldFilePath == "" && e.OriginalProfileInfo != nil && e.OriginalProfileInfo.Path != "" {
			oldFilePath = e.OriginalProfileInfo.Path
		}
		if oldFilePath == "" {
			oldFilePath = filepath.Join(baseDir, shared.BuildProfileFileName(e.OriginalProfileName))
		}

		if err := os.Remove(oldFilePath); err != nil {
			ui.PrintWarning(fmt.Sprintf(consts.ProfileWarnSavedButDeleteOldFailedFmt, newFileName, oldFilePath, err))
		}
		ui.PrintSuccess(fmt.Sprintf(consts.ProfileSuccessSavedRenamedFmt, newFileName, shared.BuildProfileFileName(e.OriginalProfileName)))
		ui.PrintInfo(consts.ProfileMsgConfigSavedAtPrefix + newFilePath)
		return nil
	}

	if err := fsops.WriteFile(newFilePath, encryptedContent); err != nil {
		return fmt.Errorf(consts.ProfileErrWriteConfigFailedFmt, err)
	}

	ui.PrintSuccess(fmt.Sprintf(consts.ProfileSuccessSavedSafelyFmt, newFileName))
	ui.PrintInfo(consts.ProfileMsgConfigSavedAtPrefix + newFilePath)
	return nil
}
