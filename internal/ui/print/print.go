// File : internal/ui/print/print.go
// Deskripsi : Print helper (header/info/warn/error) untuk UI facade
// Author : Hadiyatna Muflihun
// Tanggal : 5 Januari 2026
// Last Modified : 5 Januari 2026

package print

import (
	"fmt"
	"sfDBTools/internal/domain"
	applog "sfDBTools/internal/services/log"
	"sfDBTools/internal/ui/progress"
	legacyui "sfDBTools/pkg/ui"
)

func PrintHeader(title string) {
	legacyui.PrintHeader(title)
}

func PrintSubHeader(title string) {
	legacyui.PrintSubHeader(title)
}

func PrintInfo(msg string) {
	legacyui.PrintInfo(msg)
}

func PrintWarn(msg string) {
	legacyui.PrintWarning(msg)
}

func PrintWarning(msg string) {
	legacyui.PrintWarning(msg)
}

func PrintError(msg string) {
	legacyui.PrintError(msg)
}

// PrintSuccess dipakai luas di codebase existing, jadi tetap disediakan sebagai facade.
func PrintSuccess(msg string) {
	legacyui.PrintSuccess(msg)
}

func PrintSeparator() {
	legacyui.PrintSeparator()
}

func PrintDashedSeparator() {
	legacyui.PrintDashedSeparator()
}

func PrintAppHeader(title string) {
	legacyui.Headers(title)
}

func PrintFilterStats(stats *domain.FilterStats, context string, logger applog.Logger) {
	legacyui.DisplayFilterStats(stats, context, logger)
}

// Println menambah satu baris kosong.
func Println() {
	progress.RunWithSpinnerSuspended(func() {
		fmt.Println()
	})
}
