// File : cmd/cmd_restore/cmd_restore_custom.go
// Deskripsi : Command untuk restore custom dari SFCola account detail
// Author : Hadiyatna Muflihun
// Tanggal : 24 Desember 2025

package restorecmd

import (
	"sfdbtools/internal/app/restore"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/cli/runner"

	"github.com/spf13/cobra"
)

// CmdRestoreCustom adalah command untuk restore custom (SFCola account detail)
var CmdRestoreCustom = &cobra.Command{
	Use:   "custom",
	Short: "Restore custom (paste account detail → provision DB+users → restore DB & DMART)",
	Run: func(cmd *cobra.Command, args []string) {
		runner.Run(cmd, func() error {
			_ = restore.ExecuteRestoreCustomCommand(cmd, appdeps.Deps)
			return nil
		})
	},
}

func init() {
	flags.AddRestoreCustomFlags(CmdRestoreCustom)
}
