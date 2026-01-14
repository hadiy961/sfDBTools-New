// File : internal/app/profile/display/password.go
// Deskripsi : Reveal password (interactive) + verifikasi key
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package display

import (
	"sfdbtools/internal/app/profile/helpers"
	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/internal/ui/table"

	"github.com/AlecAivazis/survey/v2"
)

func (d *Displayer) revealPasswordConfirmAndShow(orig *domain.ProfileInfo) {
	if orig.Path == "" {
		print.PrintWarning(consts.ProfileDisplayNoFileForVerify)
		return
	}

	key, err := prompt.AskPassword(consts.ProfileDisplayVerifyKeyPrompt, survey.Required)
	if err != nil {
		print.PrintWarning(consts.ProfileDisplayVerifyKeyFailedPrefix + err.Error())
		return
	}
	if key == "" {
		print.PrintWarning(consts.ProfileDisplayNoKeyProvided)
		return
	}

	realPw, err := helpers.LoadProfilePasswordFromPath(d.ConfigDir, orig.Path, key)
	if err != nil {
		print.PrintWarning(consts.ProfileDisplayInvalidKeyOrCorrupt)
		return
	}
	display := shared.DisplayValueOrNotSet(realPw)

	print.PrintSubHeader(consts.ProfileDisplayRevealedPasswordTitle)
	table.Render(
		[]string{consts.ProfileDisplayTableHeaderNo, consts.ProfileDisplayTableHeaderField, consts.ProfileDisplayTableHeaderValue},
		[][]string{{"1", consts.ProfileLabelDBPassword, display}},
	)
}
