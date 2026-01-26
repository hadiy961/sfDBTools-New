// File : internal/app/profile/executor/save.go
// Deskripsi : Simpan profile ke file (terenkripsi)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 25 Januari 2026

package executor

import (
	"crypto/subtle"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/app/profile/helpers/keys"
	"sfdbtools/internal/app/profile/merger"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/fsops"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/prompt"
)

func (e *Executor) SaveProfile(mode string) error {
	isInteractive := e.isInteractiveMode()

	skipConnTest := false
	if e.State != nil {
		if importOpts, ok := e.State.ImportOptions(); ok && importOpts != nil {
			// Import melakukan conn-test di tahap validasi (default ON) dan butuh opsi skip untuk automation.
			if importOpts.SkipConnTest || importOpts.ConnTestDone {
				skipConnTest = true
			}
		}
	}

	var baseDir string
	var originalAbsPath string
	if mode == consts.ProfileSaveModeEdit && e.State.ProfileInfo.Path != "" && filepath.IsAbs(e.State.ProfileInfo.Path) {
		originalAbsPath = e.State.ProfileInfo.Path
		baseDir = filepath.Dir(e.State.ProfileInfo.Path)
	} else {
		baseDir = e.ConfigDir
	}

	if !fsops.DirExists(baseDir) {
		if err := fsops.CreateDirIfNotExist(baseDir); err != nil {
			return fmt.Errorf(consts.ProfileErrCreateConfigDirFailedFmt, err)
		}
	}
	if !skipConnTest {
		e.Log.Info("Menghubungkan ke database target, sebelum menyimpan profile...")

		if c, err := profileconn.ConnectWithProfile(e.Config, e.State.ProfileInfo, consts.DefaultInitialDatabase); err != nil {
			if !isInteractive {
				return err
			}
			info := profileconn.DescribeConnectError(e.Config, err)
			e.Log.Warn(info.Title)
			if strings.TrimSpace(info.Detail) != "" {
				e.Log.Warn("Detail (ringkas): " + info.Detail)
			}
			for _, h := range info.Hints {
				e.Log.Info("Hint: " + h)
			}
			continueAnyway, askErr := prompt.Confirm(consts.ProfileSavePromptContinueDespiteDBFail, false)
			if askErr != nil {
				return validation.HandleInputError(askErr)
			}
			if !continueAnyway {
				return validation.ErrConnectionFailedRetry
			}
			e.Log.Warn(consts.ProfileSaveWarnSavingWithInvalidConn)
		} else {
			c.Close()
			if !isInteractive {
				e.Log.Info(consts.ProfileLogDBConnectionValid)
			}
		}
	}

	if e.Ops == nil {
		return fmt.Errorf(consts.ProfileErrFormatINIUnavailable)
	}
	iniContent := e.Ops.FormatConfigToINI()

	key, _, err := keys.ResolveProfileEncryptionKey(e.State.ProfileInfo.EncryptionKey, isInteractive)
	if err != nil {
		return fmt.Errorf(consts.ProfileErrEncryptionKeyUnavailableFmt, err)
	}
	e.State.ProfileInfo.EncryptionKey = strings.TrimSpace(key)
	if mode == consts.ProfileSaveModeEdit && isInteractive && strings.EqualFold(strings.TrimSpace(e.State.ProfileInfo.EncryptionSource), "env") {
		confirmKey, err := prompt.PromptPassword(consts.ProfileSaveVerifyKeyPrompt)
		if err != nil {
			return validation.HandleInputError(err)
		}
		// Constant-time comparison untuk prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(confirmKey)), []byte(e.State.ProfileInfo.EncryptionKey)) != 1 {
			return fmt.Errorf(consts.ProfileSaveVerifyKeyMismatch)
		}
	}

	encryptedContent, err := crypto.EncryptData([]byte(iniContent), []byte(e.State.ProfileInfo.EncryptionKey))
	if err != nil {
		return fmt.Errorf(consts.ProfileErrEncryptConfigFailedFmt, err)
	}

	if mode == consts.ProfileSaveModeEdit {
		if editOpts, ok := e.State.EditOptions(); ok && editOpts != nil {
			if strings.TrimSpace(editOpts.NewName) != "" {
				if err := validation.ValidateProfileName(editOpts.NewName); err != nil {
					return err
				}
				e.State.ProfileInfo.Name = editOpts.NewName
			}
		}
	}
	if err := validation.ValidateProfileName(e.State.ProfileInfo.Name); err != nil {
		return err
	}

	e.State.ProfileInfo.Name = profileconn.TrimProfileSuffix(e.State.ProfileInfo.Name)
	newFileName := merger.BuildProfileFileName(e.State.ProfileInfo.Name)
	newFilePath := filepath.Join(baseDir, newFileName)

	if mode == consts.ProfileSaveModeEdit && e.State.OriginalProfileName != "" && e.State.OriginalProfileName != e.State.ProfileInfo.Name {
		if err := fsops.WriteFile(newFilePath, encryptedContent); err != nil {
			return fmt.Errorf(consts.ProfileErrWriteNewConfigFailedFmt, err)
		}

		oldFilePath := originalAbsPath
		if oldFilePath == "" && e.State.OriginalProfileInfo != nil && e.State.OriginalProfileInfo.Path != "" {
			oldFilePath = e.State.OriginalProfileInfo.Path
		}
		if oldFilePath == "" {
			oldFilePath = filepath.Join(baseDir, merger.BuildProfileFileName(e.State.OriginalProfileName))
		}

		if err := os.Remove(oldFilePath); err != nil {
			e.Log.Warn(fmt.Sprintf(consts.ProfileWarnSavedButDeleteOldFailedFmt, newFileName, oldFilePath, err))
		}
		e.Log.Info(fmt.Sprintf(consts.ProfileSuccessSavedRenamedFmt, newFileName, merger.BuildProfileFileName(e.State.OriginalProfileName)))
		e.Log.Info(consts.ProfileMsgConfigSavedAtPrefix + newFilePath)
		return nil
	}

	if err := fsops.WriteFile(newFilePath, encryptedContent); err != nil {
		return fmt.Errorf(consts.ProfileErrWriteConfigFailedFmt, err)
	}

	e.Log.Info(consts.ProfileMsgConfigSavedAtPrefix + newFilePath)
	return nil
}
