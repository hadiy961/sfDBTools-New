// File : internal/app/profile/selector.go
// Deskripsi : Pemilihan profile secara interaktif (shared untuk show/edit)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 15 Januari 2026

package profile

import (
	"fmt"
	"sfdbtools/internal/app/profile/helpers/loader"
	"sfdbtools/internal/shared/consts"
)

func (s *executorOps) promptSelectExistingConfig() error {
	info, originalName, snapshot, err := loader.SelectExistingDBConfigWithSnapshot(
		s.Config.ConfigDir.DatabaseProfile,
		consts.ProfileWizardPromptSelectExistingConfig,
	)
	if err != nil {
		return err
	}

	s.State.ProfileInfo = info
	s.State.OriginalProfileName = originalName
	s.State.OriginalProfileInfo = snapshot

	s.Log.Debug(fmt.Sprintf(consts.ProfileLogConfigLoadedFromFmt, info.Path, info.Name))
	return nil
}
