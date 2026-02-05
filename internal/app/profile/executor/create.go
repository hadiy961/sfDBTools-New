// File : internal/profile/executor/create.go
// Deskripsi : Eksekusi pembuatan profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 15 Januari 2026

package executor

import (
	profiledisplay "sfdbtools/internal/app/profile/display"
	profilevalidation "sfdbtools/internal/app/profile/validation"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/shared/validation"
)

// CreateProfile membuat profile database baru dengan validation dan connection test.
//
// Flow:
//  1. Interactive mode: Run wizard untuk collect input dari user
//  2. Non-interactive mode: Validate input dari flags
//  3. Check name uniqueness (via Ops.CheckConfigurationNameUnique)
//  4. Display profile summary (untuk review)
//  5. Save profile (dengan connection test)
//  6. Retry mechanism jika connection test atau save gagal
//
// Interactive Mode:
//   - Wizard menampilkan prompts untuk DB info dan SSH tunnel
//   - User dapat review dan confirm sebelum save
//   - Validation errors ditangani dengan retry option
//   - Connection test failure dapat di-override dengan confirmation
//
// Non-Interactive Mode:
//   - Semua input harus disediakan via flags
//   - Validation errors langsung return error (no retry)
//   - Connection test failure langsung return error (no override)
//   - Quiet mode: minimal output (cocok untuk automation)
//
// Validation:
//   - Profile name: format, uniqueness
//   - Database info: host, port, user, password
//   - SSH tunnel: host, port, auth (password atau identity file)
//
// Retry Behavior:
//   - Validation error: prompt retry dari awal (wizard restart)
//   - Connection test error: prompt retry save only (skip wizard)
//   - Name conflict: prompt retry dengan name change
//
// Returns:
//   - nil: Profile created successfully
//   - error: Creation failed (validation, save, connection test, dll)
//
// Special Errors:
//   - validation.ErrUserCancelled: User cancelled wizard
//   - validation.ErrConnectionFailedRetry: User chose not to save after connection failure
//   - profileerrors.ErrProfileExists: Name conflict (handled by uniqueness check)
func (e *Executor) CreateProfile() error {
	isInteractive := e.isInteractiveMode()
	if !isInteractive {
		e.Log.Info(consts.ProfileLogCreateStarted)
	}

	// Jika retry save karena koneksi DB invalid, kita tidak ingin restart wizard dari awal.
	skipWizard := false

	for {
		createOpts, _ := e.State.CreateOptions()
		if !skipWizard && createOpts != nil && createOpts.Interactive {
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
			if !runtimecfg.IsQuiet() {
				profiledisplay.DisplayProfileDetails(e.ConfigDir, e.State)
			}
		}
		skipWizard = false

		if err := e.Ops.CheckConfigurationNameUnique(consts.ProfileModeCreate); err != nil {
			// print.PrintError(err.Error())
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
