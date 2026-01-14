// File : internal/app/profile/display/displayer.go
// Deskripsi : Tampilan detail profil (show/create/edit summary)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 14 Januari 2026

package display

import (
	"fmt"

	profilehelper "sfdbtools/internal/app/profile/helpers"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/internal/ui/table"

	"github.com/AlecAivazis/survey/v2"
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

func (d *Displayer) printShowDetails() {
	orig := d.State.OriginalProfileInfo
	if orig == nil {
		print.PrintInfo(consts.ProfileDisplayNoConfigLoaded)
		return
	}

	pwState := consts.ProfileDisplayStateNotSet
	pwState = shared.DisplayStateSetOrNotSet(orig.DBInfo.Password)

	rows := [][]string{
		{"1", consts.ProfileDisplayFieldName, orig.Name},
		{"2", consts.ProfileDisplayFieldFilePath, orig.Path},
		{"3", consts.ProfileDisplayFieldHost, orig.DBInfo.Host},
		{"4", consts.ProfileDisplayFieldPort, fmt.Sprintf("%d", orig.DBInfo.Port)},
		{"5", consts.ProfileDisplayFieldUser, orig.DBInfo.User},
		{"6", consts.ProfileDisplayFieldPassword, pwState},
		{"7", consts.ProfileDisplayFieldSSHTunnel, fmt.Sprintf("%v", orig.SSHTunnel.Enabled)},
		{"8", consts.ProfileDisplayFieldFileSize, orig.Size},
		{"9", consts.ProfileDisplayFieldLastModified, fmt.Sprintf("%v", orig.LastModified)},
	}
	if orig.SSHTunnel.Enabled {
		rows = append(rows, []string{"10", consts.ProfileLabelSSHHost, orig.SSHTunnel.Host})
		rows = append(rows, []string{"11", consts.ProfileLabelSSHUser, orig.SSHTunnel.User})
		rows = append(rows, []string{"12", consts.ProfileLabelSSHPort, fmt.Sprintf("%d", orig.SSHTunnel.Port)})
		sshPwState := shared.DisplayStateSetOrNotSet(orig.SSHTunnel.Password)
		rows = append(rows, []string{"13", consts.ProfileLabelSSHPassword, sshPwState})
	}

	table.Render([]string{consts.ProfileDisplayTableHeaderNo, consts.ProfileDisplayTableHeaderField, consts.ProfileDisplayTableHeaderValue}, rows)

	if d.State.ProfileShow != nil && d.State.ProfileShow.RevealPassword && d.State.ProfileShow.Interactive {
		d.revealPasswordConfirmAndShow(orig)
	}
}

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

	realPw, err := profilehelper.LoadProfilePasswordFromPath(d.ConfigDir, orig.Path, key)
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

func (d *Displayer) printCreateSummary() {
	if d.State.ProfileInfo == nil {
		print.PrintInfo(consts.ProfileDisplayNoProfileInfo)
		return
	}
	rows := [][]string{
		{"1", consts.ProfileDisplayFieldName, d.State.ProfileInfo.Name},
		{"2", consts.ProfileDisplayFieldHost, d.State.ProfileInfo.DBInfo.Host},
		{"3", consts.ProfileDisplayFieldPort, fmt.Sprintf("%d", d.State.ProfileInfo.DBInfo.Port)},
		{"4", consts.ProfileDisplayFieldUser, d.State.ProfileInfo.DBInfo.User},
	}

	pwState := shared.DisplayStateSetOrNotSet(d.State.ProfileInfo.DBInfo.Password)
	rows = append(rows, []string{"5", consts.ProfileDisplayFieldPassword, pwState})

	sshState := consts.ProfileDisplaySSHDisabled
	if d.State.ProfileInfo.SSHTunnel.Enabled {
		sshState = consts.ProfileDisplaySSHEnabled
	}
	rows = append(rows, []string{"6", consts.ProfileDisplayFieldSSHTunnel, sshState})

	if d.State.ProfileInfo.SSHTunnel.Enabled {
		rows = append(rows, []string{"7", consts.ProfileLabelSSHHost, d.State.ProfileInfo.SSHTunnel.Host})
		sshPwState := shared.DisplayStateSetOrNotSet(d.State.ProfileInfo.SSHTunnel.Password)
		rows = append(rows, []string{"8", consts.ProfileLabelSSHPassword, sshPwState})
	}

	table.Render([]string{consts.ProfileDisplayTableHeaderNo, consts.ProfileDisplayTableHeaderField, consts.ProfileDisplayTableHeaderValue}, rows)
}

func (d *Displayer) printChangeSummary() {
	orig := d.State.OriginalProfileInfo
	if orig == nil || d.State.ProfileInfo == nil {
		print.PrintInfo(consts.ProfileDisplayNoChangeInfo)
		return
	}

	rows := [][]string{}
	idx := 1

	pwState := func(pw string) string { return shared.DisplayStateSetOrNotSet(pw) }

	if orig.Name != d.State.ProfileInfo.Name {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileDisplayFieldName, orig.Name, d.State.ProfileInfo.Name})
		idx++
	}
	if orig.DBInfo.Host != d.State.ProfileInfo.DBInfo.Host {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileDisplayFieldHost, orig.DBInfo.Host, d.State.ProfileInfo.DBInfo.Host})
		idx++
	}
	if orig.DBInfo.Port != d.State.ProfileInfo.DBInfo.Port {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileDisplayFieldPort, fmt.Sprintf("%d", orig.DBInfo.Port), fmt.Sprintf("%d", d.State.ProfileInfo.DBInfo.Port)})
		idx++
	}
	if orig.DBInfo.User != d.State.ProfileInfo.DBInfo.User {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileDisplayFieldUser, orig.DBInfo.User, d.State.ProfileInfo.DBInfo.User})
		idx++
	}
	if orig.DBInfo.Password != d.State.ProfileInfo.DBInfo.Password {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileDisplayFieldPassword, pwState(orig.DBInfo.Password), pwState(d.State.ProfileInfo.DBInfo.Password)})
		idx++
	}
	if orig.SSHTunnel.Enabled != d.State.ProfileInfo.SSHTunnel.Enabled {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileDisplayFieldSSHTunnel, fmt.Sprintf("%v", orig.SSHTunnel.Enabled), fmt.Sprintf("%v", d.State.ProfileInfo.SSHTunnel.Enabled)})
		idx++
	}
	if orig.SSHTunnel.Host != d.State.ProfileInfo.SSHTunnel.Host {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileLabelSSHHost, orig.SSHTunnel.Host, d.State.ProfileInfo.SSHTunnel.Host})
		idx++
	}
	if orig.SSHTunnel.Password != d.State.ProfileInfo.SSHTunnel.Password {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileLabelSSHPassword, pwState(orig.SSHTunnel.Password), pwState(d.State.ProfileInfo.SSHTunnel.Password)})
		idx++
	}

	if len(rows) == 0 {
		print.PrintInfo(consts.ProfileDisplayNoChangesDetected)
		return
	}

	table.Render([]string{consts.ProfileDisplayTableHeaderNo, consts.ProfileDisplayTableHeaderField, consts.ProfileDisplayTableHeaderBefore, consts.ProfileDisplayTableHeaderAfter}, rows)
}

// DisplayProfileDetails shows profile details
func DisplayProfileDetails(configDir string, state *profilemodel.ProfileState) {
	d := &Displayer{
		ConfigDir: configDir,
		State:     state,
	}
	d.DisplayProfileDetails()
}
