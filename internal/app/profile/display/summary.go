// File : internal/app/profile/display/summary.go
// Deskripsi : Summary untuk create/edit profile
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package display

import (
	"fmt"

	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/table"
)

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

	pwState := displayStateSetOrNotSet(d.State.ProfileInfo.DBInfo.Password)
	rows = append(rows, []string{"5", consts.ProfileDisplayFieldPassword, pwState})

	sshState := consts.ProfileDisplaySSHDisabled
	if d.State.ProfileInfo.SSHTunnel.Enabled {
		sshState = consts.ProfileDisplaySSHEnabled
	}
	rows = append(rows, []string{"6", consts.ProfileDisplayFieldSSHTunnel, sshState})

	if d.State.ProfileInfo.SSHTunnel.Enabled {
		rows = append(rows, []string{"7", consts.ProfileLabelSSHHost, d.State.ProfileInfo.SSHTunnel.Host})
		sshPwState := displayStateSetOrNotSet(d.State.ProfileInfo.SSHTunnel.Password)
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

	pwState := func(pw string) string { return displayStateSetOrNotSet(pw) }

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
