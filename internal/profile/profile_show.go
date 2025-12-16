package profile

import (
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"
)

// ShowProfile menampilkan detail profil yang ada.
func (s *Service) ShowProfile() error {
	ui.Headers("Display User Profile Details")

	// 1. Pastikan flag --file diisi

	// Guard against nil or empty flags
	if s.ProfileShow == nil || strings.TrimSpace(s.ProfileShow.Path) == "" {
		// Simpan nilai RevealPassword sebelum inisialisasi ulang
		var revealPassword bool
		if s.ProfileShow != nil {
			revealPassword = s.ProfileShow.RevealPassword
		}

		if err := s.promptSelectExistingConfig(); err != nil {
			return err
		}
		// Setelah user memilih file, set ProfileShow dengan informasi yang benar
		if s.ProfileShow == nil {
			s.ProfileShow = &types.ProfileShowOptions{}
		}
		s.ProfileShow.Path = s.ProfileInfo.Path
		// Restore nilai RevealPassword yang mungkin sudah diset dari command line
		s.ProfileShow.RevealPassword = revealPassword
	} else {
		// Jika user memberikan --file, resolve path dan muat snapshot
		abs, name, err := helper.ResolveConfigPath(s.ProfileShow.Path)
		if err != nil {
			return err
		}
		// Coba muat snapshot dari path
		if !fsops.PathExists(abs) {
			return fmt.Errorf("file konfigurasi tidak ditemukan: %s", abs)
		}
		// Muat snapshot dari path
		s.ProfileInfo.Name = name
		if err := s.loadSnapshotFromPath(abs); err != nil {
			// Tetap tampilkan metadata minimum jika ada, dan informasikan errornya
			s.Log.Warn("Gagal memuat isi detail konfigurasi: " + err.Error())
		}
	}

	// Pastikan ada file yang dimuat
	if s.OriginalProfileInfo == nil || s.OriginalProfileInfo.Path == "" {
		return fmt.Errorf("tidak ada snapshot konfigurasi untuk ditampilkan")
	}

	// 2. Pastikan file ada
	if !fsops.PathExists(s.OriginalProfileInfo.Path) {
		return fmt.Errorf("file konfigurasi tidak ditemukan: %s", s.OriginalProfileInfo.Path)
	}

	// 3. Tampilkan detail
	s.ProfileInfo.Path = s.OriginalProfileInfo.Path
	// Salin DBInfo dari snapshot agar dapat digunakan untuk uji koneksi
	if s.OriginalProfileInfo != nil {
		s.ProfileInfo.DBInfo = s.OriginalProfileInfo.DBInfo
	}
	// Uji koneksi database dan tampilkan status
	if err := database.ConnectionTest(&s.ProfileInfo.DBInfo, s.Log); err != nil {
		return nil
	}

	s.DisplayProfileDetails()

	return nil
}
