// File : internal/app/profile/helpers/profile_parse.go
// Deskripsi : Utility untuk load dan parse profil terenkripsi
// Author : Hadiyatna Muflihun
// Tanggal : 5 Desember 2025
// Last Modified : 14 Januari 2026
package helpers

import (
	"sfdbtools/internal/app/profile/helpers/parser"
	"sfdbtools/internal/domain"
)

// LoadAndParseProfile membaca file terenkripsi, mendapatkan kunci (jika tidak diberikan),
// mendekripsi, dan mem-parsing isi INI menjadi ProfileInfo (tanpa metadata file).
func LoadAndParseProfile(absPath string, key string) (*domain.ProfileInfo, error) {
	return parser.LoadAndParseProfile(absPath, key)
}
