// File : internal/ui/table/table.go
// Deskripsi : Renderer tabel untuk UI facade
// Author : Hadiyatna Muflihun
// Tanggal : 5 Januari 2026
// Last Modified : 5 Januari 2026

package table

import legacyui "sfDBTools/pkg/ui"

// Render merender tabel ke stdout.
func Render(headers []string, rows [][]string) {
	legacyui.FormatTable(headers, rows)
}
