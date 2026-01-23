package cryptocmd

import (
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/cli/parsing"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/ui/print"

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
		print.PrintAppHeader("Decrypt Text Tools")

		// Validasi password aplikasi terlebih dahulu
		if err := crypto.ValidateApplicationPassword(); err != nil {
			appdeps.Deps.Logger.Error("Autentikasi gagal: " + err.Error())
			return
		}

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
