// File : internal/profile/executor/show.go
// Deskripsi : Eksekusi tampilkan profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026
package executor

import (
	"fmt"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/table"
	"sfdbtools/pkg/consts"
	"sfdbtools/pkg/fsops"
	"sfdbtools/pkg/helper"
	profilehelper "sfdbtools/internal/app/profile/helpers"
	"sfdbtools/pkg/validation"
	"strings"
)

func (e *Executor) ShowProfile() error {
	print.PrintAppHeader(consts.ProfileUIHeaderShow)
	isInteractive := e.isInteractiveMode()

	if !isInteractive {
		if e.ProfileShow == nil || strings.TrimSpace(e.ProfileShow.Path) == "" {
			return fmt.Errorf(consts.ProfileErrNonInteractiveProfileFlagRequired)
		}
		if strings.TrimSpace(e.ProfileInfo.EncryptionKey) == "" {
			return fmt.Errorf(
				consts.ProfileErrNonInteractiveProfileKeyRequiredFmt,
				consts.ENV_TARGET_PROFILE_KEY,
				consts.ENV_SOURCE_PROFILE_KEY,
				validation.ErrNonInteractive,
			)
		}
	}

	if e.ProfileShow == nil || strings.TrimSpace(e.ProfileShow.Path) == "" {
		var revealPassword bool
		if e.ProfileShow != nil {
			revealPassword = e.ProfileShow.RevealPassword
		}

		if e.PromptSelectExistingConfig == nil {
			return fmt.Errorf(consts.ProfileErrPromptSelectorUnavailable)
		}
		if err := e.PromptSelectExistingConfig(); err != nil {
			return err
		}
		if e.ProfileShow == nil {
			e.ProfileShow = &profilemodel.ProfileShowOptions{}
		}
		e.ProfileShow.Path = e.ProfileInfo.Path
		e.ProfileShow.RevealPassword = revealPassword
	} else {
		abs, name, err := helper.ResolveConfigPath(e.ProfileShow.Path)
		if err != nil {
			return err
		}
		if !fsops.PathExists(abs) {
			return fmt.Errorf(consts.ProfileErrConfigFileNotFoundFmt, abs)
		}
		e.ProfileInfo.Name = name
		if e.LoadSnapshotFromPath != nil {
			snap, err := e.LoadSnapshotFromPath(abs)
			if err != nil {
				if e.Log != nil {
					e.Log.Warn(fmt.Sprintf(consts.ProfileLogLoadConfigDetailsFailedFmt, err))
				}
				return err
			}
			e.OriginalProfileInfo = snap
		}
	}

	if e.OriginalProfileInfo == nil || e.OriginalProfileInfo.Path == "" {
		return fmt.Errorf(consts.ProfileErrNoSnapshotToShow)
	}
	if !fsops.PathExists(e.OriginalProfileInfo.Path) {
		return fmt.Errorf(consts.ProfileErrConfigFileNotFoundFmt, e.OriginalProfileInfo.Path)
	}

	e.ProfileInfo.Path = e.OriginalProfileInfo.Path
	if e.OriginalProfileInfo != nil {
		e.ProfileInfo.DBInfo = e.OriginalProfileInfo.DBInfo
	}

	if c, err := profilehelper.ConnectWithProfile(e.ProfileInfo, consts.DefaultInitialDatabase); err != nil {
		print.PrintWarning(consts.ProfileWarnDBConnectFailedPrefix + err.Error())
	} else {
		c.Close()
	}

	// Non-interaktif: --reveal-password tidak boleh prompt.
	// Fail-fast jika key salah/corrupt agar scripting mendapat exit code non-zero.
	if e.ProfileShow != nil && e.ProfileShow.RevealPassword && !isInteractive {
		if strings.TrimSpace(e.ProfileInfo.EncryptionKey) == "" {
			return fmt.Errorf(
				consts.ProfileErrNonInteractiveProfileKeyRequiredFmt,
				consts.ENV_TARGET_PROFILE_KEY,
				consts.ENV_SOURCE_PROFILE_KEY,
				validation.ErrNonInteractive,
			)
		}
		info, err := profilehelper.ResolveAndLoadProfile(profilehelper.ProfileLoadOptions{
			ConfigDir:      e.ConfigDir,
			ProfilePath:    e.OriginalProfileInfo.Path,
			ProfileKey:     e.ProfileInfo.EncryptionKey,
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

	if e.DisplayProfileDetails != nil {
		e.DisplayProfileDetails()
	}
	return nil
}
