// File : internal/app/profile/shared/connect_preflight.go
// Deskripsi : (DEPRECATED) Facade preflight koneksi profile
// Author : Hadiyatna Muflihun
// Tanggal : 9 Januari 2026
// Last Modified : 14 Januari 2026

package shared

import (
	"time"

	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/domain"
)

func ValidateConnectPreflight(profile *domain.ProfileInfo) error {
	return profileconn.ValidateConnectPreflight(profile)
}

func ProfileConnectTimeout() time.Duration {
	return profileconn.ProfileConnectTimeout()
}

// EffectiveDBInfo mengembalikan DBInfo yang efektif untuk koneksi.
// Jika SSH tunnel aktif dan sudah memiliki ResolvedLocalPort, koneksi diarahkan ke localhost.
func EffectiveDBInfo(profile *domain.ProfileInfo) domain.DBInfo {
	return profileconn.EffectiveDBInfo(profile)
}
