package dbusercmd

import (
	dbuser "sfdbtools/internal/app/dbuser"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/runner"

	"github.com/spf13/cobra"
)

var CmdDBUserApply = &cobra.Command{
	Use:     "apply",
	Aliases: []string{"restore", "import"},
	Short:   "Apply (restore) file SQL users/grants ke target server",
	Run: func(cmd *cobra.Command, args []string) {
		runner.Run(cmd, func() error {
			return dbuser.ExecuteApply(cmd, appdeps.Deps)
		})
	},
}

func init() {
	dbuser.AddApplyFlags(CmdDBUserApply)
}
