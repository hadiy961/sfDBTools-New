package cryptocmd

import "github.com/spf13/cobra"

// CmdCryptoMain adalah perintah induk untuk operasi kriptografi umum
var CmdCryptoMain = &cobra.Command{
	Use:     "crypto",
	Aliases: []string{"enc"},
	Short:   "Perintah utilitas: encrypt/decrypt dan base64",
	Long: `Kumpulan perintah utilitas untuk enkripsi/dekripsi (AES kompatibel OpenSSL) dan base64.
Gunakan 'sfdbtools crypto <sub-command> --help' untuk detail tiap perintah.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Subcommands didefinisikan di file terpisah dan di-add di init masing-masing file.
}
