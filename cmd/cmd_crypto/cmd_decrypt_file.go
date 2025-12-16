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
	decFileInPath  string
	decFileOutPath string
	decFileKey     string
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
		lg := types.Deps.Logger
		quiet := cryptohelper.SetupQuietMode(lg)

		// Password authentication
		cryptoauth.MustValidatePassword()

		if !quiet {
			ui.Headers("Decrypt File")
		}

		// Interactive mode untuk file paths jika tidak ada flag
		inputPath, err := cryptohelper.GetFilePathOrInteractive(decFileInPath, "📂 Masukkan path file terenkripsi yang akan didekripsi:", true)
		if err != nil {
			lg.Errorf("Gagal mendapatkan input file: %v", err)
			return
		}

		outputPath, err := cryptohelper.GetFilePathOrInteractive(decFileOutPath, "💾 Masukkan path file output hasil dekripsi:", false)
		if err != nil {
			lg.Errorf("Gagal mendapatkan output file: %v", err)
			return
		}

		// Dapatkan key dari flag/env/prompt
		key, _, err := helper.ResolveEncryptionKey(decFileKey, consts.ENV_ENCRYPTION_KEY)
		if err != nil {
			lg.Errorf("Gagal mendapatkan encryption key: %v", err)
			return
		}

		// Support pipeline via '-' for stdin/stdout
		if inputPath == "-" || outputPath == "-" {
			// Read all from input (stdin or file), decrypt, then write to stdout/file
			var data []byte
			if inputPath == "-" {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					lg.Errorf("Gagal membaca stdin: %v", err)
					return
				}
				data = b
			} else {
				b, err := os.ReadFile(inputPath)
				if err != nil {
					lg.Errorf("Gagal membaca file input: %v", err)
					return
				}
				data = b
			}

			plain, err := encrypt.DecryptAES(data, []byte(key))
			if err != nil {
				lg.Errorf("Gagal mendekripsi data: %v", err)
				return
			}

			if outputPath == "-" {
				if _, err := os.Stdout.Write(plain); err != nil {
					lg.Errorf("Gagal menulis stdout: %v", err)
					return
				}
			} else {
				if err := os.WriteFile(outputPath, plain, 0600); err != nil {
					lg.Errorf("Gagal menulis file output: %v", err)
					return
				}
				lg.Infof("✓ File didekripsi: %s → %s", inputPath, outputPath)
			}
		} else {
			if err := encrypt.DecryptFile(inputPath, outputPath, []byte(key)); err != nil {
				lg.Errorf("Gagal mendekripsi file: %v", err)
				return
			}
			lg.Infof("✓ File didekripsi: %s → %s", inputPath, outputPath)
		}
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdDecryptFile)
	CmdDecryptFile.Flags().StringVarP(&decFileInPath, "in", "i", "", "Path file input terenkripsi (wajib)")
	CmdDecryptFile.Flags().StringVarP(&decFileOutPath, "out", "o", "", "Path file output hasil dekripsi (wajib)")
	CmdDecryptFile.Flags().StringVarP(&decFileKey, "key", "k", "", "Encryption key (opsional, jika kosong pakai env atau prompt)")
}
