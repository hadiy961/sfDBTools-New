// File : internal/profile/select_existing.go
// Deskripsi : Pemilihan profile secara interaktif (shared untuk show/edit)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 4 Januari 2026

package profile

import (
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	profilehelper "sfDBTools/pkg/helper/profile"
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
	s.OriginalProfileInfo = &types.ProfileInfo{
		Path:         info.Path,
		Name:         info.Name,
		DBInfo:       info.DBInfo,
		SSHTunnel:    info.SSHTunnel,
		Size:         info.Size,
		LastModified: info.LastModified,
	}

	s.Log.Debug(fmt.Sprintf(consts.ProfileLogConfigLoadedFromFmt, info.Path, info.Name))
	return nil
}
