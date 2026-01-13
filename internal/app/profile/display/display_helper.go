// File : internal/app/profile/display/display_helper.go
// Deskripsi : Display helper function (extracted for P2)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package display

import (
	profilemodel "sfdbtools/internal/app/profile/model"
)

// DisplayProfileDetails shows profile details
func DisplayProfileDetails(configDir string, state *profilemodel.ProfileState) {
	d := &Displayer{
		ConfigDir: configDir,
		State:     state,
	}
	d.DisplayProfileDetails()
}
