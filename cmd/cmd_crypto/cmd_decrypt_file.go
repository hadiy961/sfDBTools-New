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
	decFileInPath  string
	decFileOutPath string
	decFileKey     string
)

// CmdDecryptFile mendekripsi file input menjadi file output menggunakan passphrase
var CmdDecryptFile = &cobra.Command{
	Use:   "decrypt-file",
	Short: "Dekripsi file terenkripsi AES (OpenSSL compatible)",
	Long:  "Mendekripsi file terenkripsi dengan AES-GCM (format 'Salted__') dan menyimpan hasilnya ke file output.",
	Example: `
	# Dekripsi file dengan key dari env SFDB_ENCRYPTION_KEY (atau prompt jika kosong)
	sfdbtools crypto decrypt-file --in backup.sql.enc --out backup.sql

	# Dekripsi file dengan key eksplisit
	sfdbtools crypto decrypt-file -i data.txt.enc -o data.txt --key "mypassword"
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
			ui.Headers("Decrypt File")
		}
		if decFileInPath == "" || decFileOutPath == "" {
			fmt.Println("✗ Wajib menyertakan --in dan --out")
			_ = cmd.Help()
			return
		}

		// Dapatkan key dari flag/env/prompt
		key, _, err := helper.ResolveEncryptionKey(decFileKey, consts.ENV_ENCRYPTION_KEY)
		if err != nil {
			lg.Errorf("Gagal mendapatkan encryption key: %v", err)
			return
		}

		// Support pipeline via '-' for stdin/stdout
		if decFileInPath == "-" || decFileOutPath == "-" {
			// Read all from input (stdin or file), decrypt, then write to stdout/file
			var data []byte
			if decFileInPath == "-" {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					lg.Errorf("Gagal membaca stdin: %v", err)
					return
				}
				data = b
			} else {
				b, err := os.ReadFile(decFileInPath)
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

			if decFileOutPath == "-" {
				if _, err := os.Stdout.Write(plain); err != nil {
					lg.Errorf("Gagal menulis stdout: %v", err)
					return
				}
			} else {
				if err := os.WriteFile(decFileOutPath, plain, 0600); err != nil {
					lg.Errorf("Gagal menulis file output: %v", err)
					return
				}
				lg.Infof("✓ File didekripsi: %s → %s", decFileInPath, decFileOutPath)
			}
		} else {
			if err := encrypt.DecryptFile(decFileInPath, decFileOutPath, []byte(key)); err != nil {
				lg.Errorf("Gagal mendekripsi file: %v", err)
				return
			}
			lg.Infof("✓ File didekripsi: %s → %s", decFileInPath, decFileOutPath)
		}
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdDecryptFile)
	CmdDecryptFile.Flags().StringVarP(&decFileInPath, "in", "i", "", "Path file input terenkripsi (wajib)")
	CmdDecryptFile.Flags().StringVarP(&decFileOutPath, "out", "o", "", "Path file output hasil dekripsi (wajib)")
	CmdDecryptFile.Flags().StringVarP(&decFileKey, "key", "k", "", "Encryption key (opsional, jika kosong pakai env atau prompt)")
}
