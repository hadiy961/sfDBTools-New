package cmdcrypto

import (
	"fmt"
	"io"
	"os"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"

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
	Long:  "Mengenkripsi file input dan menyimpan hasilnya ke file output. Kompatibel dengan 'openssl enc -pbkdf2'",
	Example: `
	# Enkripsi file dengan key dari env SFDB_ENCRYPTION_KEY (atau prompt jika kosong)
	sfdbtools crypto encrypt-file --in backup.sql --out backup.sql.enc

	# Enkripsi file dengan key eksplisit
	sfdbtools crypto encrypt-file -i data.txt -o data.txt.enc --key "mypassword"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		lg := types.Deps.Logger
		// Quiet mode: don't print banners, route logs to stderr for clean stdout
		quiet := false
		if v := os.Getenv(consts.ENV_QUIET); v != "" && v != "0" && strings.ToLower(v) != "false" {
			quiet = true
			if lg != nil {
				lg.SetOutput(os.Stderr)
			}
		}
		if !quiet {
			ui.Headers("Encrypt File")
		}
		if encFileInPath == "" || encFileOutPath == "" {
			fmt.Println("✗ Wajib menyertakan --in dan --out")
			_ = cmd.Help()
			return
		}

		// Dapatkan key dari flag/env/prompt
		key, _, err := helper.ResolveEncryptionKey(encFileKey, consts.ENV_ENCRYPTION_KEY)
		if err != nil {
			lg.Errorf("Gagal mendapatkan encryption key: %v", err)
			return
		}

		// Support pipeline via '-' for stdin/stdout
		if encFileInPath == "-" || encFileOutPath == "-" {
			var reader io.Reader
			var writer io.Writer
			// Reader
			if encFileInPath == "-" {
				reader = os.Stdin
			} else {
				f, err := os.Open(encFileInPath)
				if err != nil {
					lg.Errorf("Gagal membuka file input: %v", err)
					return
				}
				defer f.Close()
				reader = f
			}
			// Writer
			if encFileOutPath == "-" {
				writer = os.Stdout
			} else {
				f, err := os.Create(encFileOutPath)
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
			if err := encrypt.EncryptFile(encFileInPath, encFileOutPath, []byte(key)); err != nil {
				lg.Errorf("Gagal mengenkripsi file: %v", err)
				return
			}
			lg.Infof("✓ File terenkripsi: %s → %s", encFileInPath, encFileOutPath)
		}
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdEncryptFile)
	CmdEncryptFile.Flags().StringVarP(&encFileInPath, "in", "i", "", "Path file input (wajib)")
	CmdEncryptFile.Flags().StringVarP(&encFileOutPath, "out", "o", "", "Path file output (wajib)")
	CmdEncryptFile.Flags().StringVarP(&encFileKey, "key", "k", "", "Encryption key (opsional, jika kosong pakai env atau prompt)")
}
