// File : internal/ui/menu/items.go
// Deskripsi : Builder menu items dari Cobra command
// Author : Hadiyatna Muflihun
// Tanggal : 23 Januari 2026
// Last Modified : 23 Januari 2026

package menu

import (
	"fmt"

	"github.com/spf13/cobra"
)

func buildMenuItems(current *cobra.Command, hasParent bool) []menuItem {
	subs := visibleSubcommands(current)
	items := make([]menuItem, 0, len(subs)+5)

	for _, c := range subs {
		items = append(items, menuItem{
			Label: buildDisplayName(c),
			Kind:  itemCommand,
			Cmd:   c,
		})
	}

	// Special navigation options
	if hasParent {
		items = append(items, menuItem{Label: "⬅️  Kembali", Kind: itemBack})
		items = append(items, menuItem{Label: "🏠  Main Menu", Kind: itemHome})
	}

	items = append(items, menuItem{Label: "❓  Bantuan (--help)", Kind: itemHelp})
	items = append(items, menuItem{Label: "🚪  Keluar", Kind: itemExit})

	return items
}

func menuLabelFor(current *cobra.Command, isRoot bool) string {
	if isRoot {
		return "Pilih menu:"
	}
	if current == nil {
		return "Pilih sub-command:"
	}
	return fmt.Sprintf("Pilih sub-command untuk %s:", current.Name())
}
