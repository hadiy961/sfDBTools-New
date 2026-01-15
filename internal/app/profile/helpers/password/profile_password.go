// File : internal/app/profile/helpers/password/profile_password.go
// Deskripsi : Helper untuk membaca password dari profile file (setelah decrypt+parse)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package password

import (
	"fmt"
	"strings"

	profileerrors "sfdbtools/internal/app/profile/errors"
	"sfdbtools/internal/app/profile/helpers/loader"
)

// LoadProfilePasswordFromPath memuat profile dari path dan mengembalikan password DB.
// Fungsi ini tidak melakukan prompt; caller bertanggung jawab menyediakan key jika dibutuhkan.
func LoadProfilePasswordFromPath(configDir string, profilePath string, profileKey string) (string, error) {
	if strings.TrimSpace(profilePath) == "" {
		return "", profileerrors.ErrProfilePathEmpty
	}
	info, err := loader.ResolveAndLoadProfile(loader.ProfileLoadOptions{
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
