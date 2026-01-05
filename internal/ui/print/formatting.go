// File : internal/ui/print/formatting.go
// Deskripsi : Fungsi utilitas untuk output format di terminal
// Author : Hadiyatna Muflihun
// Tanggal : 3 Oktober 2024
// Last Modified : 5 Januari 2026

package print

import (
	"fmt"
	"sfDBTools/internal/ui/progress"
	"sfDBTools/internal/ui/style"
	"sfDBTools/internal/ui/text"
	"sfDBTools/pkg/runtimecfg"
	"strings"
)

// PrintColoredLine prints a line with the specified color
func PrintColoredLine(msg string, color style.Color) {
	progress.RunWithSpinnerSuspended(func() {
		fmt.Println(text.Color(msg, color))
	})
}

// PrintSuccess prints success message in green
func PrintSuccess(message string) {
	PrintColoredLine("✅ "+message, style.ColorGreen)
}

// PrintWarning prints warning message in yellow
func PrintWarning(message string) {
	PrintColoredLine(message, style.ColorYellow)
}

// PrintInfo prints info message in blue
func PrintInfo(message string) {
	PrintColoredLine(message, style.ColorBlue)
}

// PrintHeader prints a header with border
func PrintHeader(title string) {
	if runtimecfg.IsQuiet() || runtimecfg.IsDaemon() {
		return
	}

	width, _, _ := GetTerminalSize()
	if width <= 0 {
		width = 80
	}

	// Calculate padding
	titleLen := len(title)
	if titleLen+4 > width {
		width = titleLen + 4
	}

	border := strings.Repeat("=", width)
	padding := (width - titleLen - 2) / 2
	leftPad := strings.Repeat(" ", padding)
	rightPad := strings.Repeat(" ", width-titleLen-2-padding)

	progress.RunWithSpinnerSuspended(func() {
		fmt.Println()
	})
	PrintColoredLine(border, style.ColorCyan)
	PrintColoredLine("|"+leftPad+title+rightPad+"|", style.ColorCyan)
	PrintColoredLine(border, style.ColorCyan)
	progress.RunWithSpinnerSuspended(func() {
		fmt.Println()
	})
}

// PrintError prints error message in red
func PrintError(message string) {
	PrintColoredLine("❌ "+message, style.ColorRed)
}

// PrintSubHeader prints a sub-header
func PrintSubHeader(title string) {
	if runtimecfg.IsQuiet() || runtimecfg.IsDaemon() {
		return
	}

	progress.RunWithSpinnerSuspended(func() {
		fmt.Println()
	})
	PrintColoredLine("📋 "+title, style.ColorBold)
	PrintDashedSeparator()
}

// PrintWarn adalah alias kompatibilitas untuk PrintWarning.
func PrintWarn(message string) {
	PrintWarning(message)
}
