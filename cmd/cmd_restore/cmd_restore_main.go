// File : cmd/cmd_restore/cmd_restore_main.go
// Deskripsi : Root command untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-16

package cmdrestore

import (
	"github.com/spf13/cobra"
)

// CmdRestore adalah root command untuk restore operations
var CmdRestore = &cobra.Command{
	Use:     "db-restore",
	Aliases: []string{"restore", "dbrestore", "db-restore", "import"},
	Short:   "Restore database dari file backup",
	Long:    `Command untuk restore database dari file backup dengan berbagai opsi.`,
}

func init() {
	// Register subcommands
	CmdRestore.AddCommand(CmdRestoreSingle)
	CmdRestore.AddCommand(CmdRestorePrimary)
}
