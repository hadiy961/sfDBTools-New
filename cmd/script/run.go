package scriptcmd

import (
	"sfdbtools/internal/app/script"
	"sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/cli/parsing"
	"sfdbtools/internal/ui/print"

	"github.com/spf13/cobra"
)

var CmdScriptRun = &cobra.Command{
	Use:   "run",
	Short: "Jalankan .sftools",
	Long:  "Mendekripsi .sftools ke folder temporary, lalu menjalankan entrypoint-nya dengan bash.",
	Example: `
# Run bundle
sfdbtools script run --file scripts/sftr_sf7_main_menu/main_menu.sftools --key "mypassword"

# Pakai flag pendek
sfdbtools script run --file scripts/sftr_sf7_main_menu/main_menu.sftools -k "mypassword"

# Key bisa dari env SFDB_SCRIPT_KEY
SFDB_SCRIPT_KEY="mypassword" sfdbtools script run --file scripts/sftr_sf7_main_menu/main_menu.sftools

# Pass-through args ke script (gunakan --)
sfdbtools script run -f tes -k "mypassword" -- --version
sfdbtools script run -f tes -k "mypassword" -- training ticket tes
sfdbtools script run -f tes -k "mypassword" -- --mode=1
sfdbtools script run -f tes -k "mypassword" -- --mode 1
`,
	Run: func(cmd *cobra.Command, args []string) {
		print.PrintAppHeader("Script Run Tools")

		opts := parsing.ParsingScriptRunOptions(cmd)
		opts.Args = args
		if err := script.ExecuteRunBundle(deps.Deps.Logger, deps.Deps.Config, opts); err != nil {
			deps.Deps.Logger.Error(err.Error())
		}
	},
}

func init() {
	CmdScriptMain.AddCommand(CmdScriptRun)
	flags.AddScriptRunFlags(CmdScriptRun)
}
