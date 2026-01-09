// File : internal/app/profile/wizard/select.go
// Deskripsi : Pemilihan profile secara interaktif
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026

package wizard

import (
	"fmt"
	profilehelper "sfdbtools/internal/app/profile/helpers"
	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/shared/consts"
)

func (r *Runner) promptSelectExistingConfig() error {
	info, err := profilehelper.ResolveAndLoadProfile(profilehelper.ProfileLoadOptions{
		ConfigDir:         r.ConfigDir,
		ProfilePath:       "",
		AllowInteractive:  true,
		InteractivePrompt: consts.ProfileWizardPromptSelectExistingConfig,
		RequireProfile:    true,
	})
	if err != nil {
		return err
	}

	r.ProfileInfo = info
	r.OriginalProfileName = info.Name
	r.OriginalProfileInfo = shared.CloneAsOriginalProfileInfo(info)

	if r.Log != nil {
		r.Log.Debug(fmt.Sprintf(consts.ProfileLogConfigLoadedFromFmt, info.Path, info.Name))
	}
	return nil
}
