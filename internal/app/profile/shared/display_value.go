// File : internal/app/profile/shared/display_value.go
// Deskripsi : Helper shared untuk formatting output display profile
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package shared

import (
	"strings"

	"sfdbtools/internal/shared/consts"
)

// DisplayValueOrNotSet mengembalikan nilai apa adanya jika terisi, atau label NotSet jika kosong.
func DisplayValueOrNotSet(value string) string {
	if strings.TrimSpace(value) == "" {
		return consts.ProfileDisplayStateNotSet
	}
	return value
}

// DisplayStateSetOrNotSet mengembalikan label Set/NotSet berdasarkan apakah value terisi.
func DisplayStateSetOrNotSet(value string) string {
	if strings.TrimSpace(value) == "" {
		return consts.ProfileDisplayStateNotSet
	}
	return consts.ProfileDisplayStateSet
}
