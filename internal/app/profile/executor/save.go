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

// SaveProfile menyimpan profile ke disk dalam format encrypted.
//
// Flow:
//  1. Determine save directory (ConfigDir atau existing file dir untuk edit)
//  2. Test connection ke database (skip untuk import atau jika sudah tested)
//  3. Format profile ke INI content
//  4. Resolve encryption key (flag/env/prompt)
//  5. Encrypt INI content dengan AES-256-GCM
//  6. Write encrypted content ke file (.cnf.enc)
//  7. Handle rename (untuk edit mode dengan name change)
//
// Parameters:
//   - mode: Save mode ("create", "edit", "clone", "import")
//
// Behavior berdasarkan mode:
//   - create: Create new file di ConfigDir
//   - edit: Overwrite existing atau create new + delete old (jika rename)
//   - clone: Create new file dengan name baru
//   - import: Batch save, skip connection test (sudah di tahap validasi)
//
// Connection Test:
//   - Default: test connection sebelum save (safety check)
//   - Skip untuk: import (jika SkipConnTest=true atau ConnTestDone=true)
//   - Interactive: prompt user untuk continue jika test gagal
//   - Non-interactive: return error jika test gagal
//
// Encryption:
//   - Key resolution: flag → env var → interactive prompt
//   - Edit mode dengan env key: verify key dengan re-prompt (security)
//   - Algorithm: AES-256-GCM (authenticated encryption)
//
// File Naming:
//   - Format: <name>.cnf.enc
//   - Name di-sanitize (trim spaces, remove extension suffix)
//   - Validation: name uniqueness check (via Executor.Ops)
//
// Rename Handling (Edit mode):
//   - Write new file dengan nama baru
//   - Delete old file
//   - Log warning jika delete old file gagal (file sudah tersimpan)
//
// Returns:
//   - nil: Save berhasil
//   - error: Save gagal (connection test failed, encryption failed, write failed, dll)
//
// Special Errors:
//   - validation.ErrConnectionFailedRetry: User chose not to continue despite connection failure
//   - validation.ErrUserCancelled: User cancelled key prompt
//   - profileerrors.ErrProfileExists: Name conflict (handled by caller)
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
