// File : cmd/crypto/env_test.go
// Deskripsi : Command untuk test encode/decode ENV (roundtrip test)
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

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// CmdEnvTest melakukan test roundtrip encode/decode untuk memverifikasi master key konsisten.
var CmdEnvTest = &cobra.Command{
	Use:   "env-test",
	Short: "Test encode/decode ENV (roundtrip test)",
	Long:  "Melakukan test roundtrip: encode plaintext lalu decode kembali untuk memverifikasi master key konsisten. Berguna untuk debugging masalah dekripsi.",
	Example: `
	# Test dengan nilai default
	sfdbtools crypto env-test

	# Test dengan nilai custom
	sfdbtools crypto env-test --text 'my-test-value'
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		print.PrintAppHeader("Env Test Tools (Roundtrip Test)")

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

		text, _ := cmd.Flags().GetString("text")
		if strings.TrimSpace(text) == "" {
			// Jika stdin adalah TTY, prompt
			if isatty.IsTerminal(os.Stdin.Fd()) {
				v, err := prompt.AskText("Masukkan plaintext untuk test (default: 'test'): ")
				if err != nil {
					return err
				}
				text = v
				if strings.TrimSpace(text) == "" {
					text = "test"
				}
			} else {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("gagal membaca stdin: %w", err)
				}
				text = string(b)
				if strings.TrimSpace(text) == "" {
					text = "test"
				}
			}
		}

		plaintext := strings.TrimSpace(text)
		if plaintext == "" {
			plaintext = "test"
		}

		fmt.Printf("Test Value: %q\n", plaintext)
		fmt.Println()

		// Step 1: Encode
		fmt.Printf("Step 1: Encoding...\n")
		encoded, err := crypto.EncodeEnvSecret(plaintext)
		if err != nil {
			return fmt.Errorf("gagal encode: %w", err)
		}
		fmt.Printf("✓ Encoded: %s\n", encoded)
		fmt.Println()

		// Step 2: Decode
		fmt.Printf("Step 2: Decoding...\n")
		decoded, wasEncrypted, err := crypto.DecodeEnvSecret(encoded)
		if err != nil {
			return fmt.Errorf("✗ Gagal decode: %w\n\nIni menunjukkan master key tidak konsisten antara encoding dan decoding!", err)
		}

		if !wasEncrypted {
			return fmt.Errorf("✗ Error: nilai yang di-encode tidak terdeteksi sebagai terenkripsi")
		}

		fmt.Printf("✓ Decoded: %q\n", decoded)
		fmt.Println()

		// Step 3: Verify
		if decoded == plaintext {
			fmt.Printf("✓✓✓ TEST PASSED: Roundtrip berhasil!\n")
			fmt.Printf("   Master key konsisten dan dapat digunakan untuk encode/decode.\n")
			return nil
		} else {
			return fmt.Errorf("✗✗✗ TEST FAILED: Nilai tidak match!\n   Expected: %q\n   Got:      %q", plaintext, decoded)
		}
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdEnvTest)
	CmdEnvTest.Flags().StringP("text", "t", "", "Plaintext untuk test (default: 'test')")
}
