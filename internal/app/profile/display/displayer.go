// File : internal/app/profile/display/displayer.go
// Deskripsi : Tampilan detail profil (show/create/edit summary)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026

package display

import (
	"fmt"

	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/internal/ui/table"
	"sfdbtools/pkg/consts"
	profilehelper "sfdbtools/pkg/helper/profile"

	"github.com/AlecAivazis/survey/v2"
)

type Displayer struct {
	ConfigDir           string
	ProfileInfo         *domain.ProfileInfo
	OriginalProfileInfo *domain.ProfileInfo
	ProfileShow         *profilemodel.ProfileShowOptions
}

func (d *Displayer) DisplayProfileDetails() {
	if d.ProfileShow != nil {
		if d.OriginalProfileInfo != nil {
			title := d.OriginalProfileInfo.Name
			if title == "" && d.ProfileInfo != nil {
				title = d.ProfileInfo.Name
			}
			print.PrintSubHeader(consts.ProfileDisplayShowPrefix + title)
			d.printShowDetails()
			return
		}
		if d.ProfileInfo != nil {
			print.PrintSubHeader(consts.ProfileDisplayShowPrefix + d.ProfileInfo.Name)
		}
		d.printCreateSummary()
		return
	}

	if d.OriginalProfileInfo != nil {
		if d.ProfileInfo != nil {
			print.PrintSubHeader(consts.ProfileMsgChangeSummaryPrefix + d.ProfileInfo.Name)
		}
		d.printChangeSummary()
		return
	}

	if d.ProfileInfo != nil {
		print.PrintSubHeader(consts.ProfileDisplayCreatePrefix + d.ProfileInfo.Name)
	}
	d.printCreateSummary()
}

func (d *Displayer) printShowDetails() {
	orig := d.OriginalProfileInfo
	if orig == nil {
		print.PrintInfo(consts.ProfileDisplayNoConfigLoaded)
		return
	}

	pwState := consts.ProfileDisplayStateNotSet
	if orig.DBInfo.Password != "" {
		pwState = consts.ProfileDisplayStateSet
	}

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
		sshPwState := consts.ProfileDisplayStateNotSet
		if orig.SSHTunnel.Password != "" {
			sshPwState = consts.ProfileDisplayStateSet
		}
		rows = append(rows, []string{"13", consts.ProfileLabelSSHPassword, sshPwState})
	}

	table.Render([]string{consts.ProfileDisplayTableHeaderNo, consts.ProfileDisplayTableHeaderField, consts.ProfileDisplayTableHeaderValue}, rows)

	if d.ProfileShow != nil && d.ProfileShow.RevealPassword && d.ProfileShow.Interactive {
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

	info, err := profilehelper.ResolveAndLoadProfile(profilehelper.ProfileLoadOptions{
		ConfigDir:      d.ConfigDir,
		ProfilePath:    orig.Path,
		ProfileKey:     key,
		RequireProfile: true,
	})
	if err != nil {
		print.PrintWarning(consts.ProfileDisplayInvalidKeyOrCorrupt)
		return
	}

	realPw := info.DBInfo.Password
	display := consts.ProfileDisplayStateNotSet
	if realPw != "" {
		display = realPw
	}

	print.PrintSubHeader(consts.ProfileDisplayRevealedPasswordTitle)
	table.Render(
		[]string{consts.ProfileDisplayTableHeaderNo, consts.ProfileDisplayTableHeaderField, consts.ProfileDisplayTableHeaderValue},
		[][]string{{"1", consts.ProfileLabelDBPassword, display}},
	)
}

func (d *Displayer) printCreateSummary() {
	if d.ProfileInfo == nil {
		print.PrintInfo(consts.ProfileDisplayNoProfileInfo)
		return
	}
	rows := [][]string{
		{"1", consts.ProfileDisplayFieldName, d.ProfileInfo.Name},
		{"2", consts.ProfileDisplayFieldHost, d.ProfileInfo.DBInfo.Host},
		{"3", consts.ProfileDisplayFieldPort, fmt.Sprintf("%d", d.ProfileInfo.DBInfo.Port)},
		{"4", consts.ProfileDisplayFieldUser, d.ProfileInfo.DBInfo.User},
	}

	pwState := consts.ProfileDisplayStateNotSet
	if d.ProfileInfo.DBInfo.Password != "" {
		pwState = consts.ProfileDisplayStateSet
	}
	rows = append(rows, []string{"5", consts.ProfileDisplayFieldPassword, pwState})

	sshState := consts.ProfileDisplaySSHDisabled
	if d.ProfileInfo.SSHTunnel.Enabled {
		sshState = consts.ProfileDisplaySSHEnabled
	}
	rows = append(rows, []string{"6", consts.ProfileDisplayFieldSSHTunnel, sshState})

	if d.ProfileInfo.SSHTunnel.Enabled {
		rows = append(rows, []string{"7", consts.ProfileLabelSSHHost, d.ProfileInfo.SSHTunnel.Host})
		sshPwState := consts.ProfileDisplayStateNotSet
		if d.ProfileInfo.SSHTunnel.Password != "" {
			sshPwState = consts.ProfileDisplayStateSet
		}
		rows = append(rows, []string{"8", consts.ProfileLabelSSHPassword, sshPwState})
	}

	table.Render([]string{consts.ProfileDisplayTableHeaderNo, consts.ProfileDisplayTableHeaderField, consts.ProfileDisplayTableHeaderValue}, rows)
}

func (d *Displayer) printChangeSummary() {
	orig := d.OriginalProfileInfo
	if orig == nil || d.ProfileInfo == nil {
		print.PrintInfo(consts.ProfileDisplayNoChangeInfo)
		return
	}

	rows := [][]string{}
	idx := 1

	pwState := func(pw string) string {
		if pw == "" {
			return consts.ProfileDisplayStateNotSet
		}
		return consts.ProfileDisplayStateSet
	}

	if orig.Name != d.ProfileInfo.Name {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileDisplayFieldName, orig.Name, d.ProfileInfo.Name})
		idx++
	}
	if orig.DBInfo.Host != d.ProfileInfo.DBInfo.Host {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileDisplayFieldHost, orig.DBInfo.Host, d.ProfileInfo.DBInfo.Host})
		idx++
	}
	if orig.DBInfo.Port != d.ProfileInfo.DBInfo.Port {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileDisplayFieldPort, fmt.Sprintf("%d", orig.DBInfo.Port), fmt.Sprintf("%d", d.ProfileInfo.DBInfo.Port)})
		idx++
	}
	if orig.DBInfo.User != d.ProfileInfo.DBInfo.User {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileDisplayFieldUser, orig.DBInfo.User, d.ProfileInfo.DBInfo.User})
		idx++
	}
	if orig.DBInfo.Password != d.ProfileInfo.DBInfo.Password {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileDisplayFieldPassword, pwState(orig.DBInfo.Password), pwState(d.ProfileInfo.DBInfo.Password)})
		idx++
	}
	if orig.SSHTunnel.Enabled != d.ProfileInfo.SSHTunnel.Enabled {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileDisplayFieldSSHTunnel, fmt.Sprintf("%v", orig.SSHTunnel.Enabled), fmt.Sprintf("%v", d.ProfileInfo.SSHTunnel.Enabled)})
		idx++
	}
	if orig.SSHTunnel.Host != d.ProfileInfo.SSHTunnel.Host {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileLabelSSHHost, orig.SSHTunnel.Host, d.ProfileInfo.SSHTunnel.Host})
		idx++
	}
	if orig.SSHTunnel.Password != d.ProfileInfo.SSHTunnel.Password {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), consts.ProfileLabelSSHPassword, pwState(orig.SSHTunnel.Password), pwState(d.ProfileInfo.SSHTunnel.Password)})
		idx++
	}

	if len(rows) == 0 {
		print.PrintInfo(consts.ProfileDisplayNoChangesDetected)
		return
	}

	table.Render([]string{consts.ProfileDisplayTableHeaderNo, consts.ProfileDisplayTableHeaderField, consts.ProfileDisplayTableHeaderBefore, consts.ProfileDisplayTableHeaderAfter}, rows)
}
