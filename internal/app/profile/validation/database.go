// File : internal/app/profile/validation/database.go
// Deskripsi : Validasi untuk DBInfo (MySQL/MariaDB)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package validation

import (
	"strings"

	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/domain"
)

// ValidateDBInfo melakukan validasi terhadap DBInfo
func ValidateDBInfo(db *domain.DBInfo) error {
	if db == nil {
		return shared.ErrDBInfoNil
	}
	if strings.TrimSpace(db.Host) == "" {
		return shared.ErrDBHostEmpty
	}
	if db.Port <= 0 || db.Port > 65535 {
		return shared.DBPortInvalidError(db.Port)
	}
	if strings.TrimSpace(db.User) == "" {
		return shared.ErrDBUserEmpty
	}
	// Password bisa kosong untuk beberapa auth method
	return nil
}
