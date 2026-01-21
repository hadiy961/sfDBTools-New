package scriptcmd

import (
	"sfdbtools/internal/app/script"
	"sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/cli/parsing"
	"sfdbtools/internal/ui/print"

	"github.com/spf13/cobra"
)

var CmdScriptEncrypt = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt satu folder script menjadi .sftools",
	Long:  "Membundle file entrypoint (mode single) atau seluruh isi foldernya (mode bundle) menjadi satu file .sftools terenkripsi.",
	Example: `
# Encrypt bundle dari entrypoint
sfdbtools script encrypt --file scripts/sftr_sf7_main_menu/main_menu.sh --key "mypassword"

# Pakai flag pendek
sfdbtools script encrypt --file scripts/sftr_sf7_main_menu/main_menu.sh -k "mypassword"

# Mode single (hanya 1 file script, tanpa scan folder)
sfdbtools script encrypt -f scripts/tools/hello.sh -m single -k "mypassword"

# Key bisa dari env SFDB_SCRIPT_KEY
SFDB_SCRIPT_KEY="mypassword" sfdbtools script encrypt --file scripts/sftr_sf7_main_menu/main_menu.sh
`,
	Run: func(cmd *cobra.Command, args []string) {
		print.PrintAppHeader("Script Encrypt Tools")

		opts := parsing.ParsingScriptEncryptOptions(cmd)
		if err := script.ExecuteEncryptBundle(deps.Deps.Logger, deps.Deps.Config, opts); err != nil {
			deps.Deps.Logger.Error(err.Error())
		}
	},
}

func init() {
	CmdScriptMain.AddCommand(CmdScriptEncrypt)
	flags.AddScriptEncryptFlags(CmdScriptEncrypt)
}
