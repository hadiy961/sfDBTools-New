package cmdcrypto

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"sfDBTools/internal/cryptoauth"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/cryptohelper"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"

	"github.com/spf13/cobra"
)

var (
	decTextInput string // base64 terenkripsi
	decTextOut   string
	decTextKey   string
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
		lg := types.Deps.Logger
		quiet := cryptohelper.SetupQuietMode(lg)

		// Password authentication
		cryptoauth.MustValidatePassword()

		// Ambil input base64
		b64, err := cryptohelper.GetInputStringOrInteractive(decTextInput, "Masukkan teks base64 terenkripsi:")
		if err != nil {
			lg.Errorf("Gagal membaca input: %v", err)
			return
		}

		encBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64))
		if err != nil {
			lg.Errorf("Input bukan base64 yang valid: %v", err)
			return
		}

		// Key
		key, _, err := helper.ResolveEncryptionKey(decTextKey, consts.ENV_ENCRYPTION_KEY)
		if err != nil {
			lg.Errorf("Gagal mendapatkan encryption key: %v", err)
			return
		}

		plain, err := encrypt.DecryptAES(encBytes, []byte(key))
		if err != nil {
			lg.Errorf("Gagal mendekripsi teks: %v", err)
			return
		}

		if strings.TrimSpace(decTextOut) == "" {
			if !quiet {
				ui.Headers("Decrypt Text")
				ui.PrintSubHeader("Decrypted text output : ")
			}
			fmt.Println(string(plain))
			if !quiet {
				ui.PrintDashedSeparator()
			}
			return
		}

		if err := os.WriteFile(decTextOut, plain, 0600); err != nil {
			lg.Errorf("Gagal menulis file output: %v", err)
			return
		}
		lg.Infof("✓ Teks didekripsi ditulis ke file: %s", decTextOut)
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdDecryptText)
	CmdDecryptText.Flags().StringVar(&decTextInput, "data", "", "Input terenkripsi base64 (kosongkan untuk baca dari stdin)")
	CmdDecryptText.Flags().StringVarP(&decTextOut, "out", "o", "", "File output plaintext (opsional)")
	CmdDecryptText.Flags().StringVarP(&decTextKey, "key", "k", "", "Encryption key (opsional, jika kosong pakai env atau prompt)")
}
