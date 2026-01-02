// File : internal/profile/display.go
// Deskripsi : Display functions untuk profile operations
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 2 Januari 2026

package profile

import (
	"fmt"
	"sfDBTools/internal/types"
	profilehelper "sfDBTools/pkg/helper/profile"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"

	"github.com/AlecAivazis/survey/v2"
)

// displayProfileOptions can be implemented if needed for showing general options
func (s *Service) displayProfileOptions() {
	s.Log.Debug("displayProfileOptions called (not implemented)")
}

func (s *Service) DisplayProfileDetails() {
	if s.ProfileShow != nil {
		if s.OriginalProfileInfo != nil {
			title := s.OriginalProfileInfo.Name
			if title == "" {
				title = s.ProfileInfo.Name
			}
			ui.PrintSubHeader("Menampilkan Profil: " + title)
			s.printShowDetails()
			return
		}
		ui.PrintSubHeader("Menampilkan Profil: " + s.ProfileInfo.Name)
		s.printCreateSummary()
		return
	}

	if s.OriginalProfileInfo != nil {
		ui.PrintSubHeader("Ringkasan Perubahan : " + s.ProfileInfo.Name)
		s.printChangeSummary()
		return
	}

	ui.PrintSubHeader("Konfigurasi Database Baru: " + s.ProfileInfo.Name)
	s.printCreateSummary()
}

func (s *Service) printCreateSummary() {
	rows := [][]string{
		{"1", "Nama", s.ProfileInfo.Name},
		{"2", "Host", s.ProfileInfo.DBInfo.Host},
		{"3", "Port", fmt.Sprintf("%d", s.ProfileInfo.DBInfo.Port)},
		{"4", "User", s.ProfileInfo.DBInfo.User},
	}
	pwState := "(not set)"
	if s.ProfileInfo.DBInfo.Password != "" {
		pwState = "(set)"
	}
	rows = append(rows, []string{"5", "Password", pwState})
	sshState := "disabled"
	if s.ProfileInfo.SSHTunnel.Enabled {
		sshState = "enabled"
	}
	rows = append(rows, []string{"6", "SSH Tunnel", sshState})
	if s.ProfileInfo.SSHTunnel.Enabled {
		rows = append(rows, []string{"7", "SSH Host", s.ProfileInfo.SSHTunnel.Host})
		sshPwState := "(not set)"
		if s.ProfileInfo.SSHTunnel.Password != "" {
			sshPwState = "(set)"
		}
		rows = append(rows, []string{"8", "SSH Password", sshPwState})
	}

	ui.FormatTable([]string{"No", "Field", "Value"}, rows)
}

func (s *Service) printChangeSummary() {
	orig := s.OriginalProfileInfo
	if orig == nil {
		ui.PrintInfo("Tidak ada informasi perubahan (tidak ada snapshot asli).")
		return
	}
	rows := [][]string{}
	idx := 1

	pwState := func(pw string) string {
		if pw == "" {
			return "(not set)"
		}
		return "(set)"
	}

	if orig.Name != s.ProfileInfo.Name {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), "Nama", orig.Name, s.ProfileInfo.Name})
		idx++
	}
	if orig.DBInfo.Host != s.ProfileInfo.DBInfo.Host {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), "Host", orig.DBInfo.Host, s.ProfileInfo.DBInfo.Host})
		idx++
	}
	if orig.DBInfo.Port != s.ProfileInfo.DBInfo.Port {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), "Port", fmt.Sprintf("%d", orig.DBInfo.Port), fmt.Sprintf("%d", s.ProfileInfo.DBInfo.Port)})
		idx++
	}
	if orig.DBInfo.User != s.ProfileInfo.DBInfo.User {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), "User", orig.DBInfo.User, s.ProfileInfo.DBInfo.User})
		idx++
	}
	if orig.DBInfo.Password != s.ProfileInfo.DBInfo.Password {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), "Password", pwState(orig.DBInfo.Password), pwState(s.ProfileInfo.DBInfo.Password)})
		idx++
	}
	if orig.SSHTunnel.Enabled != s.ProfileInfo.SSHTunnel.Enabled {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), "SSH Tunnel", fmt.Sprintf("%v", orig.SSHTunnel.Enabled), fmt.Sprintf("%v", s.ProfileInfo.SSHTunnel.Enabled)})
		idx++
	}
	if orig.SSHTunnel.Host != s.ProfileInfo.SSHTunnel.Host {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), "SSH Host", orig.SSHTunnel.Host, s.ProfileInfo.SSHTunnel.Host})
		idx++
	}
	if orig.SSHTunnel.Password != s.ProfileInfo.SSHTunnel.Password {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), "SSH Password", pwState(orig.SSHTunnel.Password), pwState(s.ProfileInfo.SSHTunnel.Password)})
		idx++
	}

	if len(rows) == 0 {
		ui.PrintInfo("Tidak ada perubahan yang terdeteksi pada konfigurasi.")
		return
	}

	ui.FormatTable([]string{"No", "Field", "Before", "After"}, rows)
}

func (s *Service) printShowDetails() {
	orig := s.OriginalProfileInfo
	if orig == nil {
		ui.PrintInfo("Tidak ada konfigurasi yang dimuat untuk ditampilkan.")
		return
	}

	pwState := "(not set)"
	if orig.DBInfo.Password != "" {
		pwState = "(set)"
	}

	rows := [][]string{
		{"1", "Nama", orig.Name},
		{"2", "File Path", orig.Path},
		{"3", "Host", orig.DBInfo.Host},
		{"4", "Port", fmt.Sprintf("%d", orig.DBInfo.Port)},
		{"5", "User", orig.DBInfo.User},
		{"6", "Password", pwState},
		{"7", "SSH Tunnel", fmt.Sprintf("%v", orig.SSHTunnel.Enabled)},
		{"8", "File Size", orig.Size},
		{"9", "Last Modified", fmt.Sprintf("%v", orig.LastModified)},
	}
	if orig.SSHTunnel.Enabled {
		rows = append(rows, []string{"10", "SSH Host", orig.SSHTunnel.Host})
		rows = append(rows, []string{"11", "SSH User", orig.SSHTunnel.User})
		rows = append(rows, []string{"12", "SSH Port", fmt.Sprintf("%d", orig.SSHTunnel.Port)})
		sshPwState := "(not set)"
		if orig.SSHTunnel.Password != "" {
			sshPwState = "(set)"
		}
		rows = append(rows, []string{"13", "SSH Password", sshPwState})
	}

	ui.FormatTable([]string{"No", "Field", "Value"}, rows)

	if s.ProfileShow != nil && s.ProfileShow.RevealPassword {
		s.revealPasswordConfirmAndShow(orig)
	}
}

func (s *Service) revealPasswordConfirmAndShow(orig *types.ProfileInfo) {
	if orig.Path == "" {
		ui.PrintWarning("Tidak ada file yang terkait untuk memverifikasi password.")
		return
	}

	key, err := input.AskPassword("Masukkan ulang encryption key untuk verifikasi: ", survey.Required)
	if err != nil {
		ui.PrintWarning("Gagal mendapatkan encryption key: " + err.Error())
		return
	}
	if key == "" {
		ui.PrintWarning("Tidak ada encryption key yang diberikan. Tidak dapat menampilkan password asli.")
		return
	}

	info, err := profilehelper.ResolveAndLoadProfile(profilehelper.ProfileLoadOptions{
		ConfigDir:      s.Config.ConfigDir.DatabaseProfile,
		ProfilePath:    orig.Path,
		ProfileKey:     key,
		RequireProfile: true,
	})
	if err != nil {
		ui.PrintWarning("Enkripsi key salah atau file rusak. Tidak dapat menampilkan password asli.")
		return
	}
	realPw := info.DBInfo.Password

	display := "(not set)"
	if realPw != "" {
		display = realPw
	}

	ui.PrintSubHeader("Revealed Password")
	ui.FormatTable([]string{"No", "Field", "Value"}, [][]string{{"1", "Database Password", display}})
}
