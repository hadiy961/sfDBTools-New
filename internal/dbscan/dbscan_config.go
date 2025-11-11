// File : internal/dbscan/dbscan_config.go
// Deskripsi : Fungsi untuk memuat dan menampilkan konfigurasi koneksi database
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 16 Oktober 2025

package dbscan

import (
	"fmt"
	"sfDBTools/pkg/profilehelper"
)

// CheckAndSelectConfigFile memeriksa file konfigurasi yang ada atau memandu pengguna untuk memilihnya.
func (s *Service) CheckAndSelectConfigFile() error {
	// Gunakan profilehelper untuk load source profile dengan interactive mode
	profile, err := profilehelper.LoadSourceProfile(
		s.ScanOptions.ProfileInfo.Path,
		s.ScanOptions.Encryption.Key,
		true, // enableInteractive - tampilkan selector jika path kosong
	)
	if err != nil {
		return fmt.Errorf("gagal load source profile: %w", err)
	}

	// Simpan ke ScanOptions (dereference pointer)
	s.ScanOptions.ProfileInfo = *profile
	return nil
}
