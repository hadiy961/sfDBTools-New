// File : internal/app/profile/display/show_formatter.go
// Deskripsi : Formatter untuk tampilan show profile
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package display

import (
	"fmt"

	"sfdbtools/internal/app/profile/formatter"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/table"
)

func (d *Displayer) printShowDetails() {
	orig := d.State.OriginalProfileInfo
	if orig == nil {
		print.PrintInfo(consts.ProfileDisplayNoConfigLoaded)
		return
	}

	pwState := formatter.DisplayStateSetOrNotSet(orig.DBInfo.Password)

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
		sshPwState := formatter.DisplayStateSetOrNotSet(orig.SSHTunnel.Password)
		rows = append(rows, []string{"13", consts.ProfileLabelSSHPassword, sshPwState})
	}

	table.Render([]string{consts.ProfileDisplayTableHeaderNo, consts.ProfileDisplayTableHeaderField, consts.ProfileDisplayTableHeaderValue}, rows)

	if d.State.ProfileShow != nil && d.State.ProfileShow.RevealPassword && d.State.ProfileShow.Interactive {
		d.revealPasswordConfirmAndShow(orig)
	}
}
