// File : internal/ui/menu/display.go
// Deskripsi : Formatting label menu
// Author : Hadiyatna Muflihun
// Tanggal : 23 Januari 2026
// Last Modified : 23 Januari 2026

package menu

import (
	"fmt"

	"github.com/spf13/cobra"
)

func buildDisplayName(cmd *cobra.Command) string {
	icon := getCommandIcon(cmd.Name())
	name := cmd.Name()

	trail := ""
	if hasVisibleSubcommands(cmd) {
		trail = " ›"
	}

	display := fmt.Sprintf("%s %s%s", icon, name, trail)
	if cmd.Short != "" {
		shortDesc := truncate(cmd.Short, 70)
		display = fmt.Sprintf("%s - %s", display, shortDesc)
	}
	return display
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
