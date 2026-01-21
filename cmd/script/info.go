package scriptcmd

import (
	"sfdbtools/internal/app/script"
	"sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/cli/parsing"
	"sfdbtools/internal/ui/print"

	"github.com/spf13/cobra"
)

var CmdScriptInfo = &cobra.Command{
	Use:   "info",
	Short: "Tampilkan info bundle .sftools",
	Long:  "Mendekripsi bundle .sftools dan membaca manifest untuk menampilkan entrypoint dan metadata.",
	Example: `
# Info bundle
sfdbtools script info -f /etc/sfDBTools/scripts/tes.sftools -k "mypassword"
`,
	Run: func(cmd *cobra.Command, args []string) {
		print.PrintAppHeader("Script Info Tools")

		opts := parsing.ParsingScriptInfoOptions(cmd)
		if err := script.ExecuteGetBundleInfo(deps.Deps.Logger, deps.Deps.Config, opts); err != nil {
			deps.Deps.Logger.Error(err.Error())
		}
	},
}

func init() {
	CmdScriptMain.AddCommand(CmdScriptInfo)
	flags.AddScriptInfoFlags(CmdScriptInfo)
}
