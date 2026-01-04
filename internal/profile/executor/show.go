// File : internal/profile/executor/show.go
// Deskripsi : Eksekusi tampilkan profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 4 Januari 2026

package executor

import (
	"fmt"
	"strings"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	profilehelper "sfDBTools/pkg/helper/profile"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

func (e *Executor) ShowProfile() error {
	ui.Headers(consts.ProfileUIHeaderShow)
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
			e.ProfileShow = &types.ProfileShowOptions{}
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
			if snap, err := e.LoadSnapshotFromPath(abs); err != nil {
				if e.Log != nil {
					e.Log.Warn(fmt.Sprintf(consts.ProfileLogLoadConfigDetailsFailedFmt, err))
				}
			} else {
				e.OriginalProfileInfo = snap
			}
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
		ui.PrintWarning(consts.ProfileWarnDBConnectFailedPrefix + err.Error())
	} else {
		c.Close()
	}

	if e.DisplayProfileDetails != nil {
		e.DisplayProfileDetails()
	}
	return nil
}
