// File : internal/app/profile/select_existing.go
// Deskripsi : Pemilihan profile secara interaktif (shared untuk show/edit)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 14 Januari 2026

package profile

import (
	"fmt"
	profilehelper "sfdbtools/internal/app/profile/helpers"
	"sfdbtools/internal/shared/consts"
)

func (s *executorOps) promptSelectExistingConfig() error {
	info, originalName, snapshot, err := profilehelper.SelectExistingDBConfigWithSnapshot(
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
