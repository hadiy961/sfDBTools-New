// File : internal/app/profile/helpers/profile_password.go
// Deskripsi : Helper untuk membaca password dari profile file (setelah decrypt+parse)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package helpers

import (
	"sfdbtools/internal/app/profile/helpers/password"
)

// LoadProfilePasswordFromPath memuat profile dari path dan mengembalikan password DB.
// Fungsi ini tidak melakukan prompt; caller bertanggung jawab menyediakan key jika dibutuhkan.
func LoadProfilePasswordFromPath(configDir string, profilePath string, profileKey string) (string, error) {
	return password.LoadProfilePasswordFromPath(configDir, profilePath, profileKey)
}
