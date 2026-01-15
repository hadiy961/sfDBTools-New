// File : internal/profile/executor/show.go
// Deskripsi : Eksekusi tampilkan profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 15 Januari 2026
package executor

import (
	"fmt"
	"strings"

	profileconn "sfdbtools/internal/app/profile/connection"
	profiledisplay "sfdbtools/internal/app/profile/display"
	"sfdbtools/internal/app/profile/helpers/loader"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/fsops"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/table"
)

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

	// Mode non-interaktif: jangan coba konek DB dan jangan print warning/hints.
	// Tujuannya agar output `profile show` stabil untuk scripting/pipeline.
	if isInteractive {
		if c, err := profileconn.ConnectWithProfile(e.Config, e.State.ProfileInfo, consts.DefaultInitialDatabase); err != nil {
			info := profileconn.DescribeConnectError(e.Config, err)
			print.PrintWarning("⚠️  " + info.Title)
			if strings.TrimSpace(info.Detail) != "" {
				print.PrintWarning("Detail (ringkas): " + info.Detail)
			}
			for _, h := range info.Hints {
				print.PrintInfo("Hint: " + h)
			}
		} else {
			c.Close()
		}
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
