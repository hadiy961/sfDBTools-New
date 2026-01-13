// File : internal/app/profile/shared/connect_effective.go
// Deskripsi : Utility untuk menghitung DBInfo efektif ketika SSH tunnel aktif (shared)
// Author : Hadiyatna Muflihun
// Tanggal : 9 Januari 2026
// Last Modified : 13 Januari 2026

package shared

import "sfdbtools/internal/domain"

// EffectiveDBInfo mengembalikan DBInfo yang efektif untuk koneksi.
// Jika SSH tunnel aktif dan sudah memiliki ResolvedLocalPort, koneksi diarahkan ke localhost.
func EffectiveDBInfo(profile *domain.ProfileInfo) domain.DBInfo {
	if profile == nil {
		return domain.DBInfo{}
	}
	info := profile.DBInfo
	if profile.SSHTunnel.Enabled && profile.SSHTunnel.ResolvedLocalPort > 0 {
		info.Host = "127.0.0.1"
		info.Port = profile.SSHTunnel.ResolvedLocalPort
	}
	return info
}
