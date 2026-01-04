// File : internal/profile/wizard/select.go
// Deskripsi : Pemilihan profile secara interaktif
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 4 Januari 2026

package wizard

import (
	"fmt"
	"sfDBTools/internal/profile/shared"
	"sfDBTools/pkg/consts"
	profilehelper "sfDBTools/pkg/helper/profile"
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
