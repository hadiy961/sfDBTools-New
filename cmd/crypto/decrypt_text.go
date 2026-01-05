package cryptocmd

import (
	appdeps "sfDBTools/internal/cli/deps"
	"sfDBTools/internal/cli/flags"
	"sfDBTools/internal/cli/parsing"
	"sfDBTools/internal/services/crypto"

	"github.com/spf13/cobra"
)

// CmdDecryptText mendekripsi teks terenkripsi base64 dan mencetak plaintext ke stdout atau file
var CmdDecryptText = &cobra.Command{
	Use:   "decrypt-text",
	Short: "Dekripsi teks terenkripsi (input base64)",
	Long:  "Mendekripsi teks terenkripsi AES-GCM yang diberikan dalam format base64 melalui --data atau stdin. Output berupa plaintext ke stdout atau file jika --out ditentukan.",
	Example: `
	# Dekripsi dari base64 via pipe
	echo -n 'BASE64DATA...' | sfdbtools crypto decrypt-text -k "mypassword"

	# Dekripsi dari flag dan simpan ke file
	sfdbtools crypto decrypt-text --data 'BASE64DATA...' -k "mypassword" --out secret.txt
	`,
	Run: func(cmd *cobra.Command, args []string) {
		opts := parsing.ParsingDecryptTextOptions(cmd)
		if err := crypto.ExecuteDecryptText(appdeps.Deps.Logger, opts); err != nil {
			appdeps.Deps.Logger.Error(err.Error())
		}
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdDecryptText)
	flags.AddDecryptTextFlags(CmdDecryptText)
}
