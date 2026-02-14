package dbusercmd

import (
	dbuser "sfdbtools/internal/app/dbuser"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/runner"

	"github.com/spf13/cobra"
)

var CmdDBUserExport = &cobra.Command{
	Use:     "export",
	Aliases: []string{"backup", "dump"},
	Short:   "Export user accounts dan/atau grants ke file SQL",
	Run: func(cmd *cobra.Command, args []string) {
		runner.Run(cmd, func() error {
			return dbuser.ExecuteExport(cmd, appdeps.Deps)
		})
	},
}

func init() {
	dbuser.AddExportFlags(CmdDBUserExport)
}
