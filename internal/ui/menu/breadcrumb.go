// File : internal/ui/menu/breadcrumb.go
// Deskripsi : Breadcrumb title untuk header menu
// Author : Hadiyatna Muflihun
// Tanggal : 23 Januari 2026
// Last Modified : 23 Januari 2026

package menu

import (
	"strings"

	"github.com/spf13/cobra"
)

func breadcrumbFromStack(stack []*cobra.Command) string {
	// Root label konsisten dengan header main.go
	base := "Main Menu"
	if len(stack) <= 1 {
		return base
	}

	parts := make([]string, 0, len(stack)-1)
	for i := 1; i < len(stack); i++ {
		if stack[i] == nil {
			continue
		}
		parts = append(parts, stack[i].Name())
	}
	if len(parts) == 0 {
		return base
	}
	return base + " > " + strings.Join(parts, " > ")
}
