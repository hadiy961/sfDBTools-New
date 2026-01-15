// File : internal/app/profile/shared/profile_connection.go
// Deskripsi : (DEPRECATED) Facade untuk koneksi profile (kompatibilitas)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package shared

import (
	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/database"
)

// ConnectWithProfile membuat koneksi database menggunakan ProfileInfo.
func ConnectWithProfile(profile *domain.ProfileInfo, initialDB string) (*database.Client, error) {
	return profileconn.ConnectWithProfile(profile, initialDB)
}

// TrimProfileSuffix menghapus suffix ekstensi profile (.cnf/.enc) dari nama jika ada.
func TrimProfileSuffix(name string) string {
	return profileconn.TrimProfileSuffix(name)
}
