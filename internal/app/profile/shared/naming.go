// File : internal/app/profile/shared/naming.go
// Deskripsi : Helper shared untuk normalisasi penamaan file profile
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package shared

import (
	"strings"

	"sfdbtools/internal/shared/consts"
)

// TrimProfileSuffix menghapus suffix ekstensi profile (.cnf/.enc) dari nama jika ada.
func TrimProfileSuffix(name string) string {
	n := strings.TrimSpace(name)
	n = strings.TrimSuffix(n, consts.ExtEnc)
	n = strings.TrimSuffix(n, consts.ExtCnf)
	return n
}
