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
	encTextInput string
	encTextOut   string
	encTextKey   string
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
		lg := types.Deps.Logger
		quiet := cryptohelper.SetupQuietMode(lg)

		// Password authentication
		cryptoauth.MustValidatePassword()

		// Ambil input
		data, err := cryptohelper.GetInputBytesOrInteractive(encTextInput, "Masukkan teks yang akan dienkripsi:")
		if err != nil {
			lg.Errorf("Gagal membaca input: %v", err)
			return
		}

		// Key
		key, _, err := helper.ResolveEncryptionKey(encTextKey, consts.ENV_ENCRYPTION_KEY)
		if err != nil {
			lg.Errorf("Gagal mendapatkan encryption key: %v", err)
			return
		}

		encBytes, err := encrypt.EncryptAES(data, []byte(key))
		if err != nil {
			lg.Errorf("Gagal mengenkripsi teks: %v", err)
			return
		}

		if strings.TrimSpace(encTextOut) == "" {
			// Cetak base64 ke stdout
			if !quiet {
				ui.Headers("Encrypt Text")
				ui.PrintSubHeader("Encrypted text output : ")
			}
			fmt.Println(base64.StdEncoding.EncodeToString(encBytes))
			if !quiet {
				ui.PrintDashedSeparator()
			}
			return
		}

		// Tulis biner ke file
		if err := os.WriteFile(encTextOut, encBytes, 0600); err != nil {
			lg.Errorf("Gagal menulis file output: %v", err)
			return
		}
		lg.Infof("✓ Teks terenkripsi ditulis ke file: %s", encTextOut)
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdEncryptText)
	CmdEncryptText.Flags().StringVar(&encTextInput, "text", "", "Teks input (kosongkan untuk baca dari stdin)")
	CmdEncryptText.Flags().StringVarP(&encTextOut, "out", "o", "", "File output biner (opsional)")
	CmdEncryptText.Flags().StringVarP(&encTextKey, "key", "k", "", "Encryption key (opsional, jika kosong pakai env atau prompt)")
}
