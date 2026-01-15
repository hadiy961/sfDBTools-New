// File : internal/app/profile/validation/database.go
// Deskripsi : Validasi untuk DBInfo (MySQL/MariaDB)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package validation

import (
	"strings"

	profileerrors "sfdbtools/internal/app/profile/errors"
	"sfdbtools/internal/domain"
)

// ValidateDBInfo melakukan validasi terhadap DBInfo
func ValidateDBInfo(db *domain.DBInfo) error {
	if db == nil {
		return profileerrors.ErrDBInfoNil
	}
	if strings.TrimSpace(db.Host) == "" {
		return profileerrors.ErrDBHostEmpty
	}
	if db.Port <= 0 || db.Port > 65535 {
		return profileerrors.DBPortInvalidError(db.Port)
	}
	if strings.TrimSpace(db.User) == "" {
		return profileerrors.ErrDBUserEmpty
	}
	// Password bisa kosong untuk beberapa auth method
	return nil
}
