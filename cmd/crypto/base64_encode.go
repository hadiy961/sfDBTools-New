package cryptocmd

import (
	appdeps "sfDBTools/internal/cli/deps"
	"sfDBTools/internal/cli/flags"
	"sfDBTools/internal/cli/parsing"
	"sfDBTools/internal/services/crypto"

	"github.com/spf13/cobra"
)

// CmdBase64Encode melakukan base64 encode
var CmdBase64Encode = &cobra.Command{
	Use:     "base64encode",
	Aliases: []string{"b64e", "b64-enc"},
	Short:   "Base64-encode data dari --text, stdin, atau mode interaktif",
	Long:    "Encode data ke format base64. Bisa dari flag --text, pipe stdin, atau mode interaktif (paste & Ctrl+D).",
	Run: func(cmd *cobra.Command, args []string) {
		opts := parsing.ParsingBase64EncodeOptions(cmd)
		if err := crypto.ExecuteBase64Encode(appdeps.Deps.Logger, opts); err != nil {
			appdeps.Deps.Logger.Error(err.Error())
		}
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdBase64Encode)
	flags.AddBase64EncodeFlags(CmdBase64Encode)
}
