// File : internal/app/profile/select_existing.go
// Deskripsi : Pemilihan profile secara interaktif (shared untuk show/edit)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026

package profile

import (
	"fmt"
	profilehelper "sfdbtools/internal/app/profile/helpers"
	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/pkg/consts"
)

func (s *Service) promptSelectExistingConfig() error {
	info, err := profilehelper.ResolveAndLoadProfile(profilehelper.ProfileLoadOptions{
		ConfigDir:         s.Config.ConfigDir.DatabaseProfile,
		ProfilePath:       "",
		AllowInteractive:  true,
		InteractivePrompt: consts.ProfileWizardPromptSelectExistingConfig,
		RequireProfile:    true,
	})
	if err != nil {
		return err
	}

	s.ProfileInfo = info
	s.OriginalProfileName = info.Name
	s.OriginalProfileInfo = shared.CloneAsOriginalProfileInfo(info)

	s.Log.Debug(fmt.Sprintf(consts.ProfileLogConfigLoadedFromFmt, info.Path, info.Name))
	return nil
}
