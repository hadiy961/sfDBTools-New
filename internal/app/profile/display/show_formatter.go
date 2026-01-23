// File : internal/app/profile/display/show_formatter.go
// Deskripsi : Formatter untuk tampilan show profile
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 21 Januari 2026

package display

import (
	"fmt"
	"strings"
	"time"

	profileconn "sfdbtools/internal/app/profile/connection"
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

	showOpts, _ := d.State.ShowOptions()
	isInteractive := showOpts != nil && showOpts.Interactive
	testSkipped := !isInteractive

	// Jalankan test koneksi (hanya untuk mode interaktif), tapi tetap tampilkan barisnya.
	var report *profileconn.ConnectionTestReport
	if isInteractive {
		report = profileconn.TestConnection(nil, orig, consts.DefaultInitialDatabase)
	} else {
		report = &profileconn.ConnectionTestReport{
			DNSResolution:  profileconn.StepResult{Status: profileconn.StepStatusSkipped, Detail: "non-interaktif"},
			TCPConnection:  profileconn.StepResult{Status: profileconn.StepStatusSkipped, Detail: "non-interaktif"},
			SSHTunnel:      profileconn.StepResult{Status: profileconn.StepStatusSkipped, Detail: "non-interaktif"},
			Authentication: profileconn.StepResult{Status: profileconn.StepStatusSkipped, Detail: "non-interaktif"},
			DBVersion:      "-",
			Healthy:        false,
		}
	}

	sshEnabled := consts.ProfileDisplayValueNo
	if orig.SSHTunnel.Enabled {
		sshEnabled = consts.ProfileDisplayValueYes
	}
	sshPortValue := "Disabled"
	if orig.SSHTunnel.Enabled {
		p := orig.SSHTunnel.Port
		if p == 0 {
			p = 22
		}
		sshPortValue = fmt.Sprintf("%d", p)
	}
	sshIdentity := consts.ProfileDisplayValueNo
	if strings.TrimSpace(orig.SSHTunnel.IdentityFile) != "" {
		sshIdentity = fmt.Sprintf("%s (%s)", consts.ProfileDisplayValueYes, orig.SSHTunnel.IdentityFile)
	}

	rows := [][]string{}
	add := func(category string, field string, value string) {
		rows = append(rows, []string{category, field, value})
	}
	addGroup := func(category string, items [][2]string) {
		for i, it := range items {
			cat := ""
			if i == 0 {
				cat = category
			}
			add(cat, it[0], it[1])
		}
	}

	addGroup(consts.ProfileDisplayCategoryProfileInfo, [][2]string{
		{consts.ProfileDisplayFieldName, orig.Name},
		{consts.ProfileDisplayFieldFilePath, orig.Path},
		{consts.ProfileDisplayFieldLastModified, fmt.Sprintf("%v", orig.LastModified)},
	})

	addGroup(consts.ProfileDisplayCategoryDBInfo, [][2]string{
		{consts.ProfileDisplayFieldHost, orig.DBInfo.Host},
		{consts.ProfileDisplayFieldPort, fmt.Sprintf("%d", orig.DBInfo.Port)},
		{consts.ProfileDisplayFieldUser, orig.DBInfo.User},
		{consts.ProfileDisplayFieldPassword, displayStateSetOrNotSet(orig.DBInfo.Password)},
	})

	addGroup(consts.ProfileDisplayCategorySSHTunnel, [][2]string{
		{"Enabled", sshEnabled},
		{consts.ProfileLabelSSHHost, displayValueOrNotSet(orig.SSHTunnel.Host)},
		{consts.ProfileLabelSSHPort, sshPortValue},
		{consts.ProfileLabelSSHUser, displayValueOrNotSet(orig.SSHTunnel.User)},
		{consts.ProfileLabelSSHPassword, displayStateSetOrNotSet(orig.SSHTunnel.Password)},
		{consts.ProfileLabelSSHIdentityFile, sshIdentity},
	})

	addGroup(consts.ProfileDisplayCategoryTestResult, [][2]string{
		{"DNS Resolution", report.DNSResolution.Display()},
		{"TCP Connection", report.TCPConnection.Display()},
		{"SSH Tunnel", report.SSHTunnel.Display()},
		{"Authentication", report.Authentication.Display()},
		{"DB Version", displayValueOrNotSet(report.DBVersion)},
	})

	health := "UNHEALTHY"
	if testSkipped {
		health = "SKIPPED"
	} else if report.Err == nil {
		health = "HEALTHY"
	}
	addGroup(consts.ProfileDisplayCategoryStatus, [][2]string{
		{"Latency", func() string {
			if testSkipped {
				return "-"
			}
			return report.TotalLatency.Round(time.Millisecond).String()
		}()},
		{"Health", health},
	})

	table.Render([]string{consts.ProfileDisplayTableHeaderCategory, consts.ProfileDisplayTableHeaderField, consts.ProfileDisplayTableHeaderValue}, rows)

	// Jika test gagal, tampilkan hint singkat setelah tabel (hanya interaktif).
	if isInteractive && report != nil && report.Err != nil {
		info := profileconn.DescribeConnectError(nil, report.Err)
		print.PrintWarning("⚠️  " + info.Title)
		if strings.TrimSpace(info.Detail) != "" {
			print.PrintWarning("Detail (ringkas): " + info.Detail)
		}
		for _, h := range info.Hints {
			print.PrintInfo("Hint: " + h)
		}
	}

	if showOpts != nil && showOpts.RevealPassword && showOpts.Interactive {
		d.revealPasswordConfirmAndShow(orig)
	}
}
