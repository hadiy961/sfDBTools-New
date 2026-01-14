// File : internal/app/profile/wizard/select.go
// Deskripsi : Pemilihan profile secara interaktif
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 14 Januari 2026

package wizard

import (
	"fmt"
	profilehelper "sfdbtools/internal/app/profile/helpers"
	"sfdbtools/internal/shared/consts"
)

func (r *Runner) promptSelectExistingConfig() error {
	info, originalName, snapshot, err := profilehelper.SelectExistingDBConfigWithSnapshot(
		r.ConfigDir,
		consts.ProfileWizardPromptSelectExistingConfig,
	)
	if err != nil {
		return err
	}

	r.State.ProfileInfo = info
	r.State.OriginalProfileName = originalName
	r.State.OriginalProfileInfo = snapshot

	r.Log.Debug(fmt.Sprintf(consts.ProfileLogConfigLoadedFromFmt, info.Path, info.Name))
	return nil
}
