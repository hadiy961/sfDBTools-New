// File : cmd/crypto/env_decode.go
// Deskripsi : Command untuk decode dan test nilai ENV terenkripsi
// Author : Hadiyatna Muflihun
// Tanggal : 30 Januari 2026
package cryptocmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// CmdEnvDecode mendekode dan menampilkan nilai ENV terenkripsi untuk testing/debugging.
var CmdEnvDecode = &cobra.Command{
	Use:   "env-decode",
	Short: "Decode nilai ENV terenkripsi untuk testing",
	Long:  "Mendekode nilai ENV terenkripsi (format prefix+payload) dan menampilkan plaintext. Berguna untuk testing dan debugging masalah dekripsi.",
	Example: `
	# Decode dari stdin
	echo -n 'SFENC:...' | sfdbtools crypto env-decode

	# Decode dari flag
	sfdbtools crypto env-decode --value 'SFENC:...'
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		print.PrintAppHeader("Env Decode Tools")

		// Set MariaDB key file path dari config jika tersedia (untuk konsistensi dengan aplikasi utama)
		if appdeps.Deps != nil && appdeps.Deps.Config != nil {
			crypto.SetMariaDBKeyFilePath(appdeps.Deps.Config.Mariadb.KeyMariaNBCFile)
		}

		// Validasi password aplikasi terlebih dahulu
		if err := crypto.ValidateApplicationPassword(); err != nil {
			return fmt.Errorf("autentikasi gagal: %w", err)
		}

		// Tampilkan informasi diagnostik
		keyFile := crypto.GetMariaDBKeyFilePath()
		fmt.Printf("MariaDB Key File Path: %s\n", keyFile)
		if stat, err := os.Stat(keyFile); err == nil {
			fmt.Printf("File Status: ✓ Exists (size: %d bytes)\n", stat.Size())
			if _, readErr := os.ReadFile(keyFile); readErr == nil {
				fmt.Printf("File Access: ✓ Readable\n")
			} else {
				fmt.Printf("File Access: ✗ Cannot read (%v)\n", readErr)
				fmt.Printf("WARNING: Master key akan menggunakan hardcoded seed saja!\n")
			}
		} else {
			fmt.Printf("File Status: ✗ Not found\n")
			fmt.Printf("WARNING: Master key akan menggunakan hardcoded seed saja!\n")
		}
		fmt.Println()

		value, _ := cmd.Flags().GetString("value")
		if strings.TrimSpace(value) == "" {
			// Jika stdin adalah TTY, prompt
			if isatty.IsTerminal(os.Stdin.Fd()) {
				v, err := prompt.AskText("Masukkan nilai ENV terenkripsi (format SFENC:...): ", prompt.WithValidator(survey.Required))
				if err != nil {
					return err
				}
				value = v
			} else {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("gagal membaca stdin: %w", err)
				}
				value = string(b)
			}
		}

		value = strings.TrimSpace(value)
		if value == "" {
			return fmt.Errorf("nilai ENV kosong")
		}

		decoded, wasEncrypted, err := crypto.DecodeEnvSecret(value)
		if err != nil {
			return fmt.Errorf("gagal mendekode: %w\n\nKemungkinan penyebab:\n  - Master key berbeda (cek akses ke MariaDB key file)\n  - Payload rusak atau tidak valid\n  - File key_maria_nbc.txt berbeda antara encoding dan decoding", err)
		}

		if !wasEncrypted {
			fmt.Printf("Input tidak terenkripsi (tidak ada prefix SFENC:), mengembalikan nilai asli:\n")
		} else {
			fmt.Printf("✓ Berhasil mendekode nilai ENV terenkripsi:\n")
		}
		fmt.Println(decoded)
		return nil
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdEnvDecode)
	CmdEnvDecode.Flags().StringP("value", "v", "", "Nilai ENV terenkripsi yang akan di-decode (jika kosong, baca dari stdin)")
}
