// File : internal/ui/text/text.go
// Deskripsi : Helper formatting text (warna, bold, dsb)
// Author : Hadiyatna Muflihun
// Tanggal : 5 Januari 2026
// Last Modified : 5 Januari 2026

package text

import "sfdbtools/internal/ui/style"

// Color memberi warna pada string.
func Color(s string, color style.Color) string {
	return string(color) + s + string(style.ColorReset)
}

// ColorText adalah alias kompatibilitas untuk Color.
func ColorText(s string, color style.Color) string {
	return Color(s, color)
}

// Bold membuat string menjadi bold.
func Bold(s string) string {
	return string(style.ColorBold) + s + string(style.ColorReset)
}
