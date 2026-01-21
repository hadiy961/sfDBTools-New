package cryptocmd

import (
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/cli/parsing"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/ui/print"

	"github.com/spf13/cobra"
)

// CmdDecryptFile mendekripsi file input menjadi file output menggunakan passphrase
var CmdDecryptFile = &cobra.Command{
	Use:   "decrypt-file",
	Short: "Dekripsi file terenkripsi AES (OpenSSL compatible)",
	Long:  "Mendekripsi file terenkripsi dengan AES-GCM (format 'Salted__') dan menyimpan hasilnya ke file output. Mendukung mode interaktif.",
	Example: `
	# Dekripsi file dengan key dari env SFDB_ENCRYPTION_KEY (atau prompt jika kosong)
	sfdbtools crypto decrypt-file --in backup.sql.enc --out backup.sql

	# Dekripsi file dengan key eksplisit
	sfdbtools crypto decrypt-file -i data.txt.enc -o data.txt --key "mypassword"
	
	# Mode interaktif (tanpa flags)
	sfdbtools crypto decrypt-file
	`,
	Run: func(cmd *cobra.Command, args []string) {
		print.PrintAppHeader("Decrypt File Tools")

		// Validasi password aplikasi terlebih dahulu
		if err := crypto.ValidateApplicationPassword(); err != nil {
			appdeps.Deps.Logger.Error("Autentikasi gagal: " + err.Error())
			return
		}

		opts := parsing.ParsingDecryptFileOptions(cmd)
		if err := crypto.ExecuteDecryptFile(appdeps.Deps.Logger, opts); err != nil {
			appdeps.Deps.Logger.Error(err.Error())
		}
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdDecryptFile)
	flags.AddDecryptFileFlags(CmdDecryptFile)
}
