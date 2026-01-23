package cryptocmd

import (
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/cli/parsing"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/ui/print"

	"github.com/spf13/cobra"
)

// CmdBase64Encode melakukan base64 encode
var CmdBase64Encode = &cobra.Command{
	Use:     "base64encode",
	Aliases: []string{"b64e", "b64-enc"},
	Short:   "Base64-encode data dari --text, stdin, atau mode interaktif",
	Long:    "Encode data ke format base64. Bisa dari flag --text, pipe stdin, atau mode interaktif (paste & Ctrl+D).",
	Run: func(cmd *cobra.Command, args []string) {
		print.PrintAppHeader("Base 64 Encode Tools")

		// Validasi password aplikasi terlebih dahulu
		if err := crypto.ValidateApplicationPassword(); err != nil {
			appdeps.Deps.Logger.Error("Autentikasi gagal: " + err.Error())
			return
		}

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
