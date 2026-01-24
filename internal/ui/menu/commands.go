// File : internal/ui/menu/commands.go
// Deskripsi : Helpers untuk filtering dan inspeksi sub-command
// Author : Hadiyatna Muflihun
// Tanggal : 23 Januari 2026
// Last Modified : 23 Januari 2026

package menu

import (
	"sort"

	"github.com/spf13/cobra"
)

func hasVisibleSubcommands(cmd *cobra.Command) bool {
	return len(visibleSubcommands(cmd)) > 0
}

func visibleSubcommands(cmd *cobra.Command) []*cobra.Command {
	if cmd == nil {
		return nil
	}

	cmds := cmd.Commands()
	out := make([]*cobra.Command, 0, len(cmds))
	for _, c := range cmds {
		if c == nil {
			continue
		}
		// Skip hidden, help, completion
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		out = append(out, c)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name() < out[j].Name()
	})
	return out
}
