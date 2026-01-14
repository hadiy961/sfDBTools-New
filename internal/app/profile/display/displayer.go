// File : internal/app/profile/display/displayer.go
// Deskripsi : Tampilan detail profil (show/create/edit summary)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 14 Januari 2026

package display

import (
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"
)

type Displayer struct {
	ConfigDir string
	State     *profilemodel.ProfileState // Shared state pointer
}

func (d *Displayer) DisplayProfileDetails() {
	if d.State.ProfileShow != nil {
		if d.State.OriginalProfileInfo != nil {
			title := d.State.OriginalProfileInfo.Name
			if title == "" && d.State.ProfileInfo != nil {
				title = d.State.ProfileInfo.Name
			}
			print.PrintSubHeader(consts.ProfileDisplayShowPrefix + title)
			d.printShowDetails()
			return
		}
		if d.State.ProfileInfo != nil {
			print.PrintSubHeader(consts.ProfileDisplayShowPrefix + d.State.ProfileInfo.Name)
		}
		d.printCreateSummary()
		return
	}

	if d.State.OriginalProfileInfo != nil {
		if d.State.ProfileInfo != nil {
			print.PrintSubHeader(consts.ProfileMsgChangeSummaryPrefix + d.State.ProfileInfo.Name)
		}
		d.printChangeSummary()
		return
	}

	if d.State.ProfileInfo != nil {
		print.PrintSubHeader(consts.ProfileDisplayCreatePrefix + d.State.ProfileInfo.Name)
	}
	d.printCreateSummary()
}

// DisplayProfileDetails shows profile details
func DisplayProfileDetails(configDir string, state *profilemodel.ProfileState) {
	d := &Displayer{
		ConfigDir: configDir,
		State:     state,
	}
	d.DisplayProfileDetails()
}
