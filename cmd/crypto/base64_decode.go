package cryptocmd

import (
	"sfDBTools/internal/crypto"
	"sfDBTools/internal/flags"
	"sfDBTools/internal/parsing"
	"sfDBTools/internal/types"

	"github.com/spf13/cobra"
)

// CmdBase64Decode melakukan base64 decode
var CmdBase64Decode = &cobra.Command{
	Use:     "base64decode",
	Aliases: []string{"b64d", "b64-dec"},
	Short:   "Base64-decode data dari --data, stdin, atau mode interaktif",
	Long:    "Decode data dari format base64. Bisa dari flag --data, pipe stdin, atau mode interaktif (paste & Ctrl+D).",
	Run: func(cmd *cobra.Command, args []string) {
		opts := parsing.ParsingBase64DecodeOptions(cmd)
		if err := crypto.ExecuteBase64Decode(types.Deps.Logger, opts); err != nil {
			types.Deps.Logger.Error(err.Error())
		}
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdBase64Decode)
	flags.AddBase64DecodeFlags(CmdBase64Decode)
}
