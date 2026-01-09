// File : internal/app/profile/executor/save.go
// Deskripsi : Simpan profile ke file (terenkripsi)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 9 Januari 2026

package executor

import (
	"crypto/subtle"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	profilehelper "sfdbtools/internal/app/profile/helpers"
	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/fsops"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
)

func (e *Executor) SaveProfile(mode string) error {
	isInteractive := e.isInteractiveMode()
	if e.Log != nil && !isInteractive {
		e.Log.Info(consts.ProfileLogStartSave)
	}

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
		info := profilehelper.DescribeConnectError(err)
		print.PrintWarning("\n⚠️  " + info.Title)
		if strings.TrimSpace(info.Detail) != "" {
			print.PrintWarning("Detail (ringkas): " + info.Detail)
		}
		for _, h := range info.Hints {
			print.PrintInfo("Hint: " + h)
		}
		continueAnyway, askErr := prompt.Confirm(consts.ProfileSavePromptContinueDespiteDBFail, false)
		if askErr != nil {
			return validation.HandleInputError(askErr)
		}
		if !continueAnyway {
			return validation.ErrConnectionFailedRetry
		}
		print.PrintWarning(consts.ProfileSaveWarnSavingWithInvalidConn)
	} else {
		c.Close()
		if e.Log != nil {
			if !isInteractive {
				e.Log.Info(consts.ProfileLogDBConnectionValid)
			}
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
	if mode == consts.ProfileSaveModeEdit && isInteractive && strings.EqualFold(strings.TrimSpace(e.ProfileInfo.EncryptionSource), "env") {
		confirmKey, err := prompt.PromptPassword(consts.ProfileSaveVerifyKeyPrompt)
		if err != nil {
			return validation.HandleInputError(err)
		}
		// Constant-time comparison untuk prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(confirmKey)), []byte(e.ProfileInfo.EncryptionKey)) != 1 {
			return fmt.Errorf(consts.ProfileSaveVerifyKeyMismatch)
		}
	}

	encryptedContent, err := crypto.EncryptData([]byte(iniContent), []byte(e.ProfileInfo.EncryptionKey))
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

	e.ProfileInfo.Name = profilehelper.TrimProfileSuffix(e.ProfileInfo.Name)
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
			print.PrintWarning(fmt.Sprintf(consts.ProfileWarnSavedButDeleteOldFailedFmt, newFileName, oldFilePath, err))
		}
		print.PrintSuccess(fmt.Sprintf(consts.ProfileSuccessSavedRenamedFmt, newFileName, shared.BuildProfileFileName(e.OriginalProfileName)))
		print.PrintInfo(consts.ProfileMsgConfigSavedAtPrefix + newFilePath)
		return nil
	}

	if err := fsops.WriteFile(newFilePath, encryptedContent); err != nil {
		return fmt.Errorf(consts.ProfileErrWriteConfigFailedFmt, err)
	}

	print.PrintSuccess(fmt.Sprintf(consts.ProfileSuccessSavedSafelyFmt, newFileName))
	print.PrintInfo(consts.ProfileMsgConfigSavedAtPrefix + newFilePath)
	return nil
}
