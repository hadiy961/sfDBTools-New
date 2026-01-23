// File : internal/crypto/exec_text.go
// Deskripsi : Executor untuk text encryption/decryption CLI commands
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026

package crypto

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"sfdbtools/internal/crypto/audit"
	"sfdbtools/internal/crypto/core"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mattn/go-isatty"
)

// ExecuteEncryptText mengenkripsi plaintext dan output base64 atau ke file.
func ExecuteEncryptText(logger applog.Logger, opts EncryptTextOptions) error {
	text := strings.TrimSpace(opts.InputText)
	var plaintext []byte
	if text != "" {
		plaintext = []byte(text)
	} else {
		// Jika stdin adalah TTY, jangan hang menunggu input tanpa prompt.
		if isatty.IsTerminal(os.Stdin.Fd()) {
			v, err := prompt.AskPassword("Masukkan plaintext yang akan di-encrypt: ", survey.Required)
			if err != nil {
				return err
			}
			if strings.TrimSpace(v) == "" {
				return fmt.Errorf("plaintext kosong, tidak ada yang di-encrypt")
			}
			plaintext = []byte(v)
		} else {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("gagal membaca stdin: %w", err)
			}
			if len(b) == 0 {
				return fmt.Errorf("plaintext kosong, tidak ada yang di-encrypt")
			}
			plaintext = b
		}
	}
	// Zero plaintext memory after use
	defer func() {
		for i := range plaintext {
			plaintext[i] = 0
		}
	}()

	keyStr, _, err := ResolveKey(opts.Key, consts.ENV_ENCRYPTION_KEY, true)
	if err != nil {
		return err
	}
	keyBytes := []byte(keyStr)
	defer func() {
		for i := range keyBytes {
			keyBytes[i] = 0
		}
	}()

	cipherBytes, err := EncryptData(plaintext, keyBytes)
	audit.LogOperation(logger, audit.OpEncryptText, "<stdin>", err == nil, err)
	if err != nil {
		return err
	}

	outPath := strings.TrimSpace(opts.OutputPath)
	if outPath != "" {
		if err := os.WriteFile(outPath, cipherBytes, core.SecureFilePermission); err != nil {
			return fmt.Errorf("gagal menulis output: %w", err)
		}
		print.PrintSubHeader("Hasil Enkripsi Teks")
		fmt.Printf("✓ Teks berhasil dienkripsi\n")
		fmt.Printf("  Output: %s\n", outPath)
		return nil
	}

	print.PrintSubHeader("Hasil Enkripsi Teks (Base64)")
	b64 := base64.StdEncoding.EncodeToString(cipherBytes)
	fmt.Println(b64)
	return nil
}

// ExecuteDecryptText mendekripsi ciphertext base64 dan output plaintext ke stdout/file.
func ExecuteDecryptText(logger applog.Logger, opts DecryptTextOptions) error {
	data := strings.TrimSpace(opts.InputData)
	var b64 string
	if data != "" {
		b64 = data
	} else {
		if isatty.IsTerminal(os.Stdin.Fd()) {
			v, err := prompt.AskText("Masukkan input base64: ", prompt.WithValidator(survey.Required))
			if err != nil {
				return err
			}
			b64 = strings.TrimSpace(v)
		} else {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("gagal membaca stdin: %w", err)
			}
			b64 = strings.TrimSpace(string(b))
		}
	}

	cipherBytes, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return fmt.Errorf("input base64 tidak valid: %w", err)
	}

	keyStr, _, err := ResolveKey(opts.Key, consts.ENV_ENCRYPTION_KEY, true)
	if err != nil {
		return err
	}
	keyBytes := []byte(keyStr)
	defer func() {
		for i := range keyBytes {
			keyBytes[i] = 0
		}
	}()

	plain, err := DecryptData(cipherBytes, keyBytes)
	audit.LogOperation(logger, audit.OpDecryptText, "<stdin>", err == nil, err)
	if err != nil {
		return err
	}
	// Zero plaintext memory after use
	defer func() {
		for i := range plain {
			plain[i] = 0
		}
	}()

	outPath := strings.TrimSpace(opts.OutputPath)
	if outPath != "" {
		if err := os.WriteFile(outPath, plain, core.SecureFilePermission); err != nil {
			return fmt.Errorf("gagal menulis output: %w", err)
		}
		print.PrintSubHeader("Hasil Dekripsi Teks")
		fmt.Printf("✓ Teks berhasil didekripsi\n")
		fmt.Printf("  Output: %s\n", outPath)
		return nil
	}

	print.PrintSubHeader("Hasil Dekripsi Teks")
	_, err = os.Stdout.Write(plain)
	return err
}
