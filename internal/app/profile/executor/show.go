// File : internal/profile/executor/show.go
// Deskripsi : Eksekusi tampilkan profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 21 Januari 2026
package executor

import (
	"fmt"
	"strings"

	profiledisplay "sfdbtools/internal/app/profile/display"
	"sfdbtools/internal/app/profile/helpers/loader"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/fsops"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/table"
)

// ShowProfile menampilkan profile details dengan optional connection test dan password reveal.
//
// Modes:
//  1. Interactive mode dengan --file: Load specific profile dan display
//  2. Interactive mode tanpa --file: List all profiles dan prompt selection
//  3. Non-interactive mode: Require --file flag (error jika tidak ada)
//
// Display Information:
//   - Profile Info: Name, Path, Last Modified
//   - Database Info: Host, Port, User, Password (masked)
//   - SSH Tunnel: Enabled, Host, Port, User, Authentication method
//   - Connection Test: DNS, TCP, SSH, Auth, DB Version (interactive only)
//   - Health Status: HEALTHY/UNHEALTHY/SKIPPED
//
// Connection Test:
//   - Interactive mode: Always test connection (live health check)
//   - Non-interactive mode: Skip test (hanya tampilkan config)
//   - Test steps: DNS resolution → TCP connection → SSH tunnel → Authentication
//   - Display latency dan error hints jika test gagal
//
// Password Reveal:
//   - Flag --reveal-password: Prompt confirmation lalu show plain password
//   - Security: Require user confirmation dengan re-enter profile key
//   - Only available in interactive mode
//
// List All Profiles:
//   - Scan ConfigDir untuk .cnf.enc files
//   - Display table dengan Name, Path, Last Modified
//   - Allow selection untuk show details
//
// Returns:
//   - nil: Display successful
//   - error: Profile not found, load failed, display error
//
// Special Errors:
//   - validation.ErrUserCancelled: User cancelled selection
//   - profileerrors.ErrNoProfilesFound: ConfigDir kosong (no profiles)
//   - profileerrors.ErrProfileNotFound: Specific profile tidak ditemukan
func (e *Executor) ShowProfile() error {
	isInteractive := e.isInteractiveMode()

	if !isInteractive {
		showOpts, ok := e.State.ShowOptions()
		if !ok || showOpts == nil || strings.TrimSpace(showOpts.Path) == "" {
			return fmt.Errorf(consts.ProfileErrNonInteractiveProfileFlagRequired)
		}
		if strings.TrimSpace(e.State.ProfileInfo.EncryptionKey) == "" {
			return fmt.Errorf(
				consts.ProfileErrNonInteractiveProfileKeyRequiredFmt,
				consts.ENV_TARGET_PROFILE_KEY,
				consts.ENV_SOURCE_PROFILE_KEY,
				validation.ErrNonInteractive,
			)
		}
	}

	showOpts, ok := e.State.ShowOptions()
	if !ok || showOpts == nil || strings.TrimSpace(showOpts.Path) == "" {
		var revealPassword bool
		if ok && showOpts != nil {
			revealPassword = showOpts.RevealPassword
		}

		if e.Ops == nil {
			return fmt.Errorf(consts.ProfileErrPromptSelectorUnavailable)
		}
		if err := e.Ops.PromptSelectExistingConfig(); err != nil {
			return err
		}
		// Pastikan show options ada dan berisi path dari profile info.
		newShow := &profilemodel.ProfileShowOptions{}
		newShow.Path = e.State.ProfileInfo.Path
		newShow.RevealPassword = revealPassword
		newShow.Interactive = e.isInteractiveMode()
		e.State.Options = newShow
		showOpts = newShow
	} else {
		abs, name, err := e.resolveProfilePath(showOpts.Path)
		if err != nil {
			return err
		}
		if !fsops.PathExists(abs) {
			return fmt.Errorf(consts.ProfileErrConfigFileNotFoundFmt, abs)
		}
		e.State.ProfileInfo.Name = name
		if e.Ops == nil {
			return fmt.Errorf(consts.ProfileErrLoadSnapshotUnavailable)
		}
		snap, err := e.Ops.LoadSnapshotFromPath(abs)
		if err != nil {
			e.Log.Warn(fmt.Sprintf(consts.ProfileLogLoadConfigDetailsFailedFmt, err))
			return err
		}
		e.State.OriginalProfileInfo = snap
	}

	if e.State.OriginalProfileInfo == nil || e.State.OriginalProfileInfo.Path == "" {
		return fmt.Errorf(consts.ProfileErrNoSnapshotToShow)
	}
	if !fsops.PathExists(e.State.OriginalProfileInfo.Path) {
		return fmt.Errorf(consts.ProfileErrConfigFileNotFoundFmt, e.State.OriginalProfileInfo.Path)
	}

	e.State.ProfileInfo.Path = e.State.OriginalProfileInfo.Path
	if e.State.OriginalProfileInfo != nil {
		e.State.ProfileInfo.DBInfo = e.State.OriginalProfileInfo.DBInfo
	}

	// Non-interaktif: --reveal-password tidak boleh prompt.
	// Fail-fast jika key salah/corrupt agar scripting mendapat exit code non-zero.
	showOpts, ok = e.State.ShowOptions()
	if ok && showOpts != nil && showOpts.RevealPassword && !isInteractive {
		if strings.TrimSpace(e.State.ProfileInfo.EncryptionKey) == "" {
			return fmt.Errorf(
				consts.ProfileErrNonInteractiveProfileKeyRequiredFmt,
				consts.ENV_TARGET_PROFILE_KEY,
				consts.ENV_SOURCE_PROFILE_KEY,
				validation.ErrNonInteractive,
			)
		}
		info, err := loader.ResolveAndLoadProfile(loader.ProfileLoadOptions{
			ConfigDir:      e.ConfigDir,
			ProfilePath:    e.State.OriginalProfileInfo.Path,
			ProfileKey:     e.State.ProfileInfo.EncryptionKey,
			RequireProfile: true,
		})
		if err != nil {
			return err
		}
		display := consts.ProfileDisplayStateNotSet
		if strings.TrimSpace(info.DBInfo.Password) != "" {
			display = info.DBInfo.Password
		}
		print.PrintSubHeader(consts.ProfileDisplayRevealedPasswordTitle)
		table.Render(
			[]string{consts.ProfileDisplayTableHeaderNo, consts.ProfileDisplayTableHeaderField, consts.ProfileDisplayTableHeaderValue},
			[][]string{{"1", consts.ProfileLabelDBPassword, display}},
		)
	}

	profiledisplay.DisplayProfileDetails(e.ConfigDir, e.State)
	return nil
}
