// File : internal/app/profile/validation/profile.go
// Deskripsi : Validasi untuk profile info
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package validation

import (
	"fmt"
	"strings"

	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/domain"
)

// ValidateProfileInfo melakukan validasi komprehensif terhadap ProfileInfo
func ValidateProfileInfo(p *domain.ProfileInfo) error {
	if p == nil {
		return shared.ErrProfileNil
	}
	if strings.TrimSpace(p.Name) == "" {
		return shared.ErrProfileNameEmpty
	}
	if err := ValidateDBInfo(&p.DBInfo); err != nil {
		return fmt.Errorf("validasi db info gagal: %w", err)
	}
	if p.SSHTunnel.Enabled {
		if strings.TrimSpace(p.SSHTunnel.Host) == "" {
			return shared.ErrSSHTunnelHostEmpty
		}
	}
	return nil
}
