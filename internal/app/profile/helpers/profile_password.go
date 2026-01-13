// File : internal/app/profile/helpers/profile_password.go
// Deskripsi : Helper untuk membaca password dari profile file (setelah decrypt+parse)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package helpers

import (
	"fmt"
	"sfdbtools/internal/app/profile/shared"
	"strings"
)

// LoadProfilePasswordFromPath memuat profile dari path dan mengembalikan password DB.
// Fungsi ini tidak melakukan prompt; caller bertanggung jawab menyediakan key jika dibutuhkan.
func LoadProfilePasswordFromPath(configDir string, profilePath string, profileKey string) (string, error) {
	if strings.TrimSpace(profilePath) == "" {
		return "", shared.ErrProfilePathEmpty
	}
	info, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ConfigDir:      configDir,
		ProfilePath:    profilePath,
		ProfileKey:     profileKey,
		RequireProfile: true,
	})
	if err != nil {
		return "", err
	}
	if info == nil {
		return "", fmt.Errorf("profile berhasil di-load tapi hasil nil")
	}
	return info.DBInfo.Password, nil
}
