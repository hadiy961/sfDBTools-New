package profile

import (
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/profilehelper"
	"sfDBTools/pkg/ui"

	"github.com/AlecAivazis/survey/v2"
)

func (s *Service) DisplayProfileDetails() {
	// Jika mode SHOW -> tampilkan detail penuh dari snapshot yang dimuat
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
		// Jika ProfileShow ada tapi OriginalProfileInfo nil
		ui.PrintSubHeader("Menampilkan Profil: " + s.ProfileInfo.Name)
		s.printCreateSummary()
		return
	}

	// Jika ada snapshot original -> ini alur EDIT: tampilkan hanya ringkasan perubahan
	if s.OriginalProfileInfo != nil {
		ui.PrintSubHeader("Ringkasan Perubahan : " + s.ProfileInfo.Name)
		s.printChangeSummary()
		return
	}

	// Default: create flow -> tampilkan ringkasan pembuatan
	ui.PrintSubHeader("Konfigurasi Database Baru: " + s.ProfileInfo.Name)
	s.printCreateSummary()

}

// printCreateSummary mencetak ringkasan konfigurasi baru.
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

	ui.FormatTable([]string{"No", "Field", "Value"}, rows)
}

// printChangeSummary compares OriginalProfileInfo and current ProfileInfo and prints a short summary.
func (s *Service) printChangeSummary() {
	orig := s.OriginalProfileInfo
	if orig == nil {
		ui.PrintInfo("Tidak ada informasi perubahan (tidak ada snapshot asli).")
		return
	}
	rows := [][]string{}
	idx := 1

	// helper to represent password state
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

	if len(rows) == 0 {
		ui.PrintInfo("Tidak ada perubahan yang terdeteksi pada konfigurasi.")
		return
	}

	// headers: No, Field, Before, After
	ui.FormatTable([]string{"No", "Field", "Before", "After"}, rows)
}

// printShowDetails mencetak seluruh detail konfigurasi dari snapshot (untuk mode show).
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
		{"8", "File Size", orig.Size},
		{"9", "Last Modified", fmt.Sprintf("%v", orig.LastModified)},
	}

	ui.FormatTable([]string{"No", "Field", "Value"}, rows)

	// Jika pengguna meminta reveal-password, lakukan konfirmasi dengan meminta
	// ulang encryption password dan coba dekripsi file. Hanya tampilkan password
	// asli jika dekripsi berhasil.
	if s.ProfileShow != nil && s.ProfileShow.RevealPassword {
		s.revealPasswordConfirmAndShow(orig)
	}
}

// revealPasswordConfirmAndShow meminta encryption password, mencoba mendekripsi file,
// dan menampilkan password jika dekripsi berhasil.
func (s *Service) revealPasswordConfirmAndShow(orig *types.ProfileInfo) {
	if orig.Path == "" {
		ui.PrintWarning("Tidak ada file yang terkait untuk memverifikasi password.")
		return
	}

	// Minta ulang encryption key
	key, err := input.AskPassword("Masukkan ulang encryption key untuk verifikasi: ", survey.Required)
	if err != nil {
		ui.PrintWarning("Gagal mendapatkan encryption key: " + err.Error())
		return
	}
	if key == "" {
		ui.PrintWarning("Tidak ada encryption key yang diberikan. Tidak dapat menampilkan password asli.")
		return
	}

	// Gunakan helper untuk memuat dan parse dengan resolver kunci
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

	// Tampilkan result dalam table kecil
	display := "(not set)"
	if realPw != "" {
		display = realPw
	}

	ui.PrintSubHeader("Revealed Password")
	ui.FormatTable([]string{"No", "Field", "Value"}, [][]string{{"1", "Database Password", display}})
}
