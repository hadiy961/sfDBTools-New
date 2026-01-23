package scriptcmd

import (
	"sfdbtools/internal/app/script"
	"sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/cli/parsing"
	"sfdbtools/internal/ui/print"

	"github.com/spf13/cobra"
)

var CmdScriptExtract = &cobra.Command{
	Use:   "extract",
	Short: "Extract bundle .sftools ke folder",
	Long:  "Mendekripsi .sftools lalu mengekstrak isinya ke out-dir. Command ini butuh proteksi ganda: key + password aplikasi.",
	Example: `
# Extract bundle untuk diedit
sfdbtools script extract -f /etc/sfDBTools/scripts/tes.sftools -o ./tes_extracted -k "mypassword"
`,
	Run: func(cmd *cobra.Command, args []string) {
		print.PrintAppHeader("Script Extract Tools")

		opts := parsing.ParsingScriptExtractOptions(cmd)
		if err := script.ExecuteExtractBundle(deps.Deps.Logger, deps.Deps.Config, opts); err != nil {
			deps.Deps.Logger.Error(err.Error())
		}
	},
}

func init() {
	CmdScriptMain.AddCommand(CmdScriptExtract)
	flags.AddScriptExtractFlags(CmdScriptExtract)
}
