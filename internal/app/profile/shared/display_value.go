// File : internal/app/profile/shared/display_value.go
// Deskripsi : (DEPRECATED) Facade display formatter
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package shared

import (
	"sfdbtools/internal/app/profile/formatter"
)

// DisplayValueOrNotSet mengembalikan nilai apa adanya jika terisi, atau label NotSet jika kosong.
func DisplayValueOrNotSet(value string) string {
	return formatter.DisplayValueOrNotSet(value)
}

// DisplayStateSetOrNotSet mengembalikan label Set/NotSet berdasarkan apakah value terisi.
func DisplayStateSetOrNotSet(value string) string {
	return formatter.DisplayStateSetOrNotSet(value)
}
