package cryptocmd

import (
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/cli/parsing"
	"sfdbtools/internal/crypto"

	"github.com/spf13/cobra"
)

// CmdEncryptFile mengenkripsi file input menjadi file output menggunakan passphrase
var CmdEncryptFile = &cobra.Command{
	Use:   "encrypt-file",
	Short: "Enkripsi file menggunakan AES (OpenSSL compatible)",
	Long:  "Mengenkripsi file input dan menyimpan hasilnya ke file output. Kompatibel dengan 'openssl enc -pbkdf2'. Mendukung mode interaktif.",
	Example: `
	# Enkripsi file dengan key dari env SFDB_ENCRYPTION_KEY (atau prompt jika kosong)
	sfdbtools crypto encrypt-file --in backup.sql --out backup.sql.enc

	# Enkripsi file dengan key eksplisit
	sfdbtools crypto encrypt-file -i data.txt -o data.txt.enc --key "mypassword"
	
	# Mode interaktif (tanpa flags)
	sfdbtools crypto encrypt-file
	`,
	Run: func(cmd *cobra.Command, args []string) {
		opts := parsing.ParsingEncryptFileOptions(cmd)
		if err := crypto.ExecuteEncryptFile(appdeps.Deps.Logger, opts); err != nil {
			appdeps.Deps.Logger.Error(err.Error())
		}
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdEncryptFile)
	flags.AddEncryptFileFlags(CmdEncryptFile)
}
