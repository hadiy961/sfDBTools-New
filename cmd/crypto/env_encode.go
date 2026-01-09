// File : cmd/crypto/env_encode.go
// Deskripsi : Command untuk encode nilai ENV terenkripsi
// Author : Hadiyatna Muflihun
// Tanggal : 6 Januari 2026
// Last Modified : 9 Januari 2026

package cryptocmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"sfdbtools/internal/crypto"
	"sfdbtools/internal/ui/prompt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// CmdEnvEncode membuat payload ENV terenkripsi format prefix:<payload>.
var CmdEnvEncode = &cobra.Command{
	Use:   "env-encode",
	Short: "Encode plaintext menjadi format prefix+payload untuk ENV",
	Long:  "Membuat payload terenkripsi untuk disimpan pada environment variable. Output menggunakan format '<prefix><payload>' (base64 URL-safe tanpa padding).",
	Example: `
	# Encode dari stdin (disarankan agar tidak masuk shell history)
	echo -n 'my-secret-key' | sfdbtools crypto env-encode

	# Encode dari flag
	sfdbtools crypto env-encode --text 'my-secret-key'
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		text, _ := cmd.Flags().GetString("text")
		if strings.TrimSpace(text) == "" {
			// Jika stdin adalah TTY, jangan hang menunggu input tanpa prompt.
			if isatty.IsTerminal(os.Stdin.Fd()) {
				v, err := prompt.AskPassword("Masukkan plaintext yang akan di-encode: ", survey.Required)
				if err != nil {
					return err
				}
				text = v
			} else {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("gagal membaca stdin: %w", err)
				}
				text = string(b)
			}
		}

		// Warning: jika key file MariaDB ada tapi tidak readable oleh user ini,
		// hasil encode akan fallback ke hardcoded saja dan bisa mismatch jika runtime nanti bisa membaca file tsb.
		keyFile := crypto.GetMariaDBKeyFilePath()
		if _, statErr := os.Stat(keyFile); statErr == nil {
			if _, readErr := os.ReadFile(keyFile); readErr != nil {
				fmt.Fprintf(os.Stderr, "WARNING: %s terdeteksi tapi tidak bisa dibaca (%v).\n", keyFile, readErr)
				fmt.Fprintf(os.Stderr, "WARNING: Pastikan proses runtime yang akan mendekripsi ENV punya akses yang konsisten, atau jalankan env-encode dengan user yang sama.\n")
			}
		}

		out, err := crypto.EncodeEnvSecret(text)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdEnvEncode)
	CmdEnvEncode.Flags().StringP("text", "t", "", "Plaintext yang akan di-encode (jika kosong, baca dari stdin)")

	// Sisipkan prefix di help/usage tanpa menyimpan prefix sebagai string literal di binary.
	CmdEnvEncode.Short = "Encode plaintext menjadi " + crypto.EncryptedPrefixForDisplay() + "<payload> untuk ENV"
}
