package scriptcmd

import "github.com/spf13/cobra"

// CmdScriptMain adalah perintah induk untuk fitur script bundle terenkripsi (.sftools)
var CmdScriptMain = &cobra.Command{
	Use:   "script",
	Short: "Encrypt & run bundled scripts (.sftools)",
	Long:  "Membungkus satu folder script menjadi file .sftools terenkripsi, dan menjalankannya via sfdbtools.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
