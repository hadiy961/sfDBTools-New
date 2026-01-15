// File : internal/app/profile/validation/profile.go
// Deskripsi : Validasi untuk profile info
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package validation

import (
	"fmt"
	"strings"

	profileerrors "sfdbtools/internal/app/profile/errors"
	"sfdbtools/internal/domain"
)

// ValidateProfileInfo melakukan validasi komprehensif terhadap ProfileInfo
func ValidateProfileInfo(p *domain.ProfileInfo) error {
	if p == nil {
		return profileerrors.ErrProfileNil
	}
	if strings.TrimSpace(p.Name) == "" {
		return profileerrors.ErrProfileNameEmpty
	}
	if err := ValidateDBInfo(&p.DBInfo); err != nil {
		return fmt.Errorf("validasi db info gagal: %w", err)
	}
	if p.SSHTunnel.Enabled {
		if strings.TrimSpace(p.SSHTunnel.Host) == "" {
			return profileerrors.ErrSSHTunnelHostEmpty
		}
	}
	return nil
}
