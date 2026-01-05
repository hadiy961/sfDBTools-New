// File : internal/app/profile/display_adapter.go
// Deskripsi : Adapter display (delegasi ke subpackage display)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026

package profile

import "sfDBTools/internal/app/profile/display"

func (s *Service) DisplayProfileDetails() {
	d := &display.Displayer{
		ConfigDir:           s.Config.ConfigDir.DatabaseProfile,
		ProfileInfo:         s.ProfileInfo,
		OriginalProfileInfo: s.OriginalProfileInfo,
		ProfileShow:         s.ProfileShow,
	}
	d.DisplayProfileDetails()
}
