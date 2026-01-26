// File : internal/app/dbcopy/helpers/profile.go
// Deskripsi : Helper functions untuk profile resolution
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
package helpers

import (
	"sfdbtools/internal/domain"
	"strings"
)

// ResolveTargetProfile menentukan target profile path dan key.
// Jika target tidak diisi, gunakan profile source sebagai default.
func ResolveTargetProfile(source *domain.ProfileInfo, targetPath, targetKey string) (string, string) {
	if source == nil {
		return strings.TrimSpace(targetPath), strings.TrimSpace(targetKey)
	}

	if strings.TrimSpace(targetPath) == "" {
		return strings.TrimSpace(source.Path), strings.TrimSpace(source.EncryptionKey)
	}

	return strings.TrimSpace(targetPath), strings.TrimSpace(targetKey)
}
