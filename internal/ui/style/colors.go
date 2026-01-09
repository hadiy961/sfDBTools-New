// File : internal/ui/style/colors.go
// Deskripsi : Konstanta style (warna/format) untuk UI facade
// Author : Hadiyatna Muflihun
// Tanggal : 5 Januari 2026
// Last Modified : 5 Januari 2026

package style

import "sfdbtools/internal/shared/consts"

// Color adalah alias string untuk warna/format ANSI.
// Di fase UI-1, nilai ini masih memetakan ke konstanta legacy di pkg/consts.
type Color = string

const (
	ColorReset  Color = consts.UIColorReset
	ColorRed    Color = consts.UIColorRed
	ColorGreen  Color = consts.UIColorGreen
	ColorYellow Color = consts.UIColorYellow
	ColorBlue   Color = consts.UIColorBlue
	ColorCyan   Color = consts.UIColorCyan
	ColorPurple Color = consts.UIColorPurple
	ColorBold   Color = consts.UIColorBold
)
