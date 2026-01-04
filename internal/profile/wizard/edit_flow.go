// File : internal/profile/wizard/edit_flow.go
// Deskripsi : Flow wizard untuk edit profile (honor flag overrides)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 4 Januari 2026

package wizard

import (
	"fmt"
	"strings"

	"sfDBTools/internal/profile/shared"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

func (r *Runner) runEditFlow() error {
	// Simpan override dari flag/env sebelum load snapshot.
	overrideDB := types.DBInfo{}
	overrideSSH := types.SSHTunnelConfig{}
	if r.ProfileInfo != nil {
		overrideDB = r.ProfileInfo.DBInfo
		overrideSSH = r.ProfileInfo.SSHTunnel
	}

	// Jika user memberikan --profile, jangan minta pilih lagi.
	target := ""
	if r.ProfileInfo != nil {
		target = strings.TrimSpace(r.ProfileInfo.Path)
	}
	if target == "" {
		target = strings.TrimSpace(r.OriginalProfileName)
	}

	if target != "" {
		absPath, name, err := helper.ResolveConfigPath(target)
		if err != nil {
			return err
		}
		if !fsops.PathExists(absPath) {
			return fmt.Errorf(consts.ProfileErrConfigFileNotFoundFmt, absPath)
		}
		if r.LoadSnapshotFromPath == nil {
			return fmt.Errorf(consts.ProfileErrLoadSnapshotUnavailable)
		}
		snap, err := r.LoadSnapshotFromPath(absPath)
		if err != nil {
			return err
		}
		r.OriginalProfileInfo = snap
		if r.ProfileInfo != nil {
			r.ProfileInfo.Path = absPath
			r.ProfileInfo.Name = name
		}
		r.OriginalProfileName = name
	} else {
		// Tidak ada --profile: pilih secara interaktif.
		if err := r.promptSelectExistingConfig(); err != nil {
			return err
		}
	}

	// Jadikan snapshot sebagai baseline, lalu apply override dari flag/env.
	shared.ApplySnapshotAsBaseline(r.ProfileInfo, r.OriginalProfileInfo)
	shared.ApplyDBOverrides(r.ProfileInfo, overrideDB)
	shared.ApplySSHOverrides(r.ProfileInfo, overrideSSH)

	hasFlagEdits := shared.HasAnyDBOverride(overrideDB) || shared.HasAnySSHOverride(overrideSSH) ||
		(r.ProfileEdit != nil && strings.TrimSpace(r.ProfileEdit.NewName) != "")

	if !hasFlagEdits {
		// Tampilkan isi profil terlebih dahulu
		prevShow := r.ProfileShow
		r.ProfileShow = &types.ProfileShowOptions{}
		if r.DisplayProfileDetails != nil {
			r.DisplayProfileDetails()
		}
		r.ProfileShow = prevShow

		action, err := input.SelectSingleFromList([]string{consts.ProfileActionEditData, consts.ProfileActionCancel}, consts.ProfilePromptAction)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if action == consts.ProfileActionCancel {
			return validation.ErrUserCancelled
		}

		if err := r.promptEditSelectedFields(); err != nil {
			return err
		}
	}

	// Jika SSH tunnel aktif tapi detail belum ada, minta via wizard.
	if err := r.promptSSHTunnelDetailsIfEnabled(); err != nil {
		return err
	}

	ui.PrintSubHeader(consts.ProfileMsgChangeSummaryPrefix + r.ProfileInfo.Name)
	return nil
}
