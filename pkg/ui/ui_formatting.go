// File : pkg/ui/ui_formatting.go
// Deskripsi : Fungsi utilitas untuk output format di terminal
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

// Colors for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

// ColorText applies color to text
func ColorText(text, color string) string {
	return color + text + ColorReset
}

// PrintColoredLine prints a line with the specified color
func PrintColoredLine(text, color string) {
	fmt.Println(ColorText(text, color))
}

// PrintSuccess prints success message in green
func PrintSuccess(message string) {
	PrintColoredLine("✅ "+message, ColorGreen)
}

// PrintWarning prints warning message in yellow
func PrintWarning(message string) {
	PrintColoredLine(message, ColorYellow)
}

// PrintInfo prints info message in blue
func PrintInfo(message string) {
	PrintColoredLine(message, ColorBlue)
}

// PrintHeader prints a header with border
func PrintHeader(title string) {
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

	fmt.Println()
	PrintColoredLine(border, ColorCyan)
	PrintColoredLine("|"+leftPad+title+rightPad+"|", ColorCyan)
	PrintColoredLine(border, ColorCyan)
	fmt.Println()
}

// PrintError prints error message in red
func PrintError(message string) {
	PrintColoredLine("❌ "+message, ColorRed)
}

// PrintSubHeader prints a sub-header
func PrintSubHeader(title string) {
	fmt.Println()
	PrintColoredLine("📋 "+title, ColorBold)
	PrintDashedSeparator()
}

// FormatTable formats data as a table using tablewriter library for better appearance
func FormatTable(headers []string, rows [][]string) {
	if len(headers) == 0 || len(rows) == 0 {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)

	// Set table headers using the correct method
	table.Header(headers)

	table.Bulk(rows)

	// Render the table
	table.Render()
}
