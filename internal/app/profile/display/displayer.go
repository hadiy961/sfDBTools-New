// File : internal/app/profile/display/displayer.go
// Deskripsi : Tampilan detail profil (show/create/edit summary)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 15 Januari 2026

package display

import (
	"strings"

	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"
)

// displayValueOrNotSet mengembalikan nilai apa adanya jika terisi, atau label NotSet jika kosong.
func displayValueOrNotSet(value string) string {
	if strings.TrimSpace(value) == "" {
		return consts.ProfileDisplayStateNotSet
	}
	return value
}

// displayStateSetOrNotSet mengembalikan label Set/NotSet berdasarkan apakah value terisi.
func displayStateSetOrNotSet(value string) string {
	if strings.TrimSpace(value) == "" {
		return consts.ProfileDisplayStateNotSet
	}
	return consts.ProfileDisplayStateSet
}

type Displayer struct {
	ConfigDir string
	State     *profilemodel.ProfileState // Shared state pointer
}

func (d *Displayer) DisplayProfileDetails() {
	if showOpts, ok := d.State.ShowOptions(); ok && showOpts != nil {
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
