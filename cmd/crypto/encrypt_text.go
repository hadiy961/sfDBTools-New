package cryptocmd

import (
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/cli/parsing"
	"sfdbtools/internal/services/crypto"

	"github.com/spf13/cobra"
)

// CmdEncryptText mengenkripsi teks dari flag atau stdin dan menuliskan hasil sebagai base64 (stdout) atau file biner
var CmdEncryptText = &cobra.Command{
	Use:   "encrypt-text",
	Short: "Enkripsi teks (output default base64 ke stdout)",
	Long:  "Mengenkripsi teks yang diberikan melalui --text atau stdin. Secara default, mencetak hasil terenkripsi dalam bentuk base64 ke stdout. Jika --out ditentukan, akan menyimpan hasil biner ke file.",
	Example: `
	# Enkripsi string menjadi base64
	echo -n 'hello' | sfdbtools crypto encrypt-text --key "mypassword"

	# Enkripsi string via flag dan simpan biner ke file
	sfdbtools crypto encrypt-text --text 'secret' -k "mypassword" --out secret.enc
	`,
	Run: func(cmd *cobra.Command, args []string) {
		opts := parsing.ParsingEncryptTextOptions(cmd)
		if err := crypto.ExecuteEncryptText(appdeps.Deps.Logger, opts); err != nil {
			appdeps.Deps.Logger.Error(err.Error())
		}
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdEncryptText)
	flags.AddEncryptTextFlags(CmdEncryptText)
}
