// File : internal/dbscan/dbscan_config.go
// Deskripsi : Fungsi untuk memuat dan menampilkan konfigurasi koneksi database
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 16 Oktober 2025

package dbscan

import (
	"fmt"
	"sfDBTools/internal/profileselect"
	"sfDBTools/pkg/helper"
	"strings"
)

// CheckAndSelectConfigFile memeriksa file konfigurasi yang ada atau memandu pengguna untuk memilihnya.
// Fungsi ini sekarang menggunakan fungsi generic dari pkg/dbconfig untuk menghindari duplikasi kode.
func (s *Service) CheckAndSelectConfigFile() error {
	// Jika user sudah memberikan path profile via flag/ENV, gunakan itu dan jangan prompt
	if strings.TrimSpace(s.ScanOptions.ProfileInfo.Path) != "" {
		absPath, name, err := helper.ResolveConfigPath(s.ScanOptions.ProfileInfo.Path)
		if err != nil {
			return fmt.Errorf("gagal memproses path konfigurasi: %w", err)
		}

		// Muat dan parse profil terenkripsi menggunakan key (jika ada)
		loaded, err := profileselect.LoadAndParseProfile(absPath, s.ScanOptions.Encryption.Key)
		if err != nil {
			return err
		}

		// Simpan kembali ke ScanOptions (pertahankan metadata yang relevan)
		s.ScanOptions.ProfileInfo.Path = absPath
		s.ScanOptions.ProfileInfo.Name = name
		s.ScanOptions.ProfileInfo.DBInfo = loaded.DBInfo
		s.ScanOptions.ProfileInfo.EncryptionSource = loaded.EncryptionSource
		return nil
	}

	// Jika tidak ada path diberikan, buka selector agar user memilih file profil
	info, err := profileselect.SelectExistingDBConfig("Pilih file konfigurasi database sumber:")
	if err != nil {
		return fmt.Errorf("gagal memilih konfigurasi database: %w", err)
	}
	// Simpan ke ScanOptions
	s.ScanOptions.ProfileInfo = info
	return nil
}
