package cmdcrypto

import (
	"io"
	"os"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/cryptoauth"
	"sfDBTools/pkg/cryptohelper"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"

	"github.com/spf13/cobra"
)

var (
	encFileInPath  string
	encFileOutPath string
	encFileKey     string
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
		lg := types.Deps.Logger
		quiet := cryptohelper.SetupQuietMode(lg)

		// Password authentication
		cryptoauth.MustValidatePassword()

		if !quiet {
			ui.Headers("Encrypt File")
		}

		// Interactive mode untuk file paths jika tidak ada flag
		inputPath, err := cryptohelper.GetFilePathOrInteractive(encFileInPath, "📂 Masukkan path file yang akan dienkripsi:", true)
		if err != nil {
			lg.Errorf("Gagal mendapatkan input file: %v", err)
			return
		}

		outputPath, err := cryptohelper.GetFilePathOrInteractive(encFileOutPath, "💾 Masukkan path file output terenkripsi:", false)
		if err != nil {
			lg.Errorf("Gagal mendapatkan output file: %v", err)
			return
		}

		// Dapatkan key dari flag/env/prompt
		key, _, err := helper.ResolveEncryptionKey(encFileKey, consts.ENV_ENCRYPTION_KEY)
		if err != nil {
			lg.Errorf("Gagal mendapatkan encryption key: %v", err)
			return
		}

		// Support pipeline via '-' for stdin/stdout
		if inputPath == "-" || outputPath == "-" {
			var reader io.Reader
			var writer io.Writer
			// Reader
			if inputPath == "-" {
				reader = os.Stdin
			} else {
				f, err := os.Open(inputPath)
				if err != nil {
					lg.Errorf("Gagal membuka file input: %v", err)
					return
				}
				defer f.Close()
				reader = f
			}
			// Writer
			if outputPath == "-" {
				writer = os.Stdout
			} else {
				f, err := os.Create(outputPath)
				if err != nil {
					lg.Errorf("Gagal membuat file output: %v", err)
					return
				}
				defer f.Close()
				writer = f
			}

			ew, err := encrypt.NewEncryptingWriter(writer, []byte(key))
			if err != nil {
				lg.Errorf("Gagal membuat encrypting writer: %v", err)
				return
			}
			if _, err := io.Copy(ew, reader); err != nil {
				lg.Errorf("Gagal mengenkripsi data: %v", err)
				return
			}
			if err := ew.Close(); err != nil {
				lg.Errorf("Gagal menutup writer: %v", err)
				return
			}
		} else {
			if err := encrypt.EncryptFile(inputPath, outputPath, []byte(key)); err != nil {
				lg.Errorf("Gagal mengenkripsi file: %v", err)
				return
			}
			lg.Infof("✓ File terenkripsi: %s → %s", inputPath, outputPath)
		}
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdEncryptFile)
	CmdEncryptFile.Flags().StringVarP(&encFileInPath, "in", "i", "", "Path file input (wajib)")
	CmdEncryptFile.Flags().StringVarP(&encFileOutPath, "out", "o", "", "Path file output (wajib)")
	CmdEncryptFile.Flags().StringVarP(&encFileKey, "key", "k", "", "Encryption key (opsional, jika kosong pakai env atau prompt)")
}
