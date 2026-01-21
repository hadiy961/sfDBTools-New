// File : internal/crypto/exec_file.go
// Deskripsi : Executor untuk file encryption/decryption CLI commands
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026

package crypto

import (
	"fmt"
	"os"
	"strings"

	"sfdbtools/internal/crypto/audit"
	cryptofile "sfdbtools/internal/crypto/file"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"

	"github.com/AlecAivazis/survey/v2"
)

// ExecuteEncryptFile mengenkripsi file dengan AES-256-GCM.
func ExecuteEncryptFile(logger applog.Logger, opts EncryptFileOptions) error {
	inPath := strings.TrimSpace(opts.InputPath)
	outPath := strings.TrimSpace(opts.OutputPath)
	if inPath == "" {
		v, err := prompt.AskText("Masukkan path file input: ", prompt.WithValidator(survey.Required))
		if err != nil {
			return err
		}
		inPath = strings.TrimSpace(v)
	}
	if outPath == "" {
		v, err := prompt.AskText("Masukkan path file output: ", prompt.WithValidator(survey.Required))
		if err != nil {
			return err
		}
		outPath = strings.TrimSpace(v)
	}

	// Validate paths setelah user input (prevent path traversal)
	if err := cryptofile.ValidatePath(inPath); err != nil {
		return fmt.Errorf("invalid input path: %w", err)
	}
	if err := cryptofile.ValidatePath(outPath); err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	// Check if input file exists
	if _, err := os.Stat(inPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file input tidak ditemukan: %s", inPath)
		}
		return fmt.Errorf("gagal mengakses file input: %w", err)
	}

	// Check if output file exists and confirm overwrite
	if _, err := os.Stat(outPath); err == nil {
		confirm, err := prompt.AskConfirm(fmt.Sprintf("File '%s' sudah ada. Timpa?", outPath), false)
		if err != nil {
			return err
		}
		if !confirm {
			return fmt.Errorf("operasi dibatalkan oleh user")
		}
	}

	keyStr, _, err := ResolveKey(opts.Key, consts.ENV_ENCRYPTION_KEY, true)
	if err != nil {
		return err
	}

	err = EncryptFile(inPath, outPath, []byte(keyStr))
	audit.LogOperation(logger, audit.OpEncryptFile, inPath, err == nil, err)
	if err != nil {
		return err
	}

	print.PrintSubHeader("Hasil Enkripsi File")
	fmt.Printf("✓ File berhasil dienkripsi\n")
	fmt.Printf("  Input:  %s\n", inPath)
	fmt.Printf("  Output: %s\n", outPath)
	return nil
}

// ExecuteDecryptFile mendekripsi file dengan AES-256-GCM.
func ExecuteDecryptFile(logger applog.Logger, opts DecryptFileOptions) error {
	inPath := strings.TrimSpace(opts.InputPath)
	outPath := strings.TrimSpace(opts.OutputPath)
	if inPath == "" {
		v, err := prompt.AskText("Masukkan path file input terenkripsi: ", prompt.WithValidator(survey.Required))
		if err != nil {
			return err
		}
		inPath = strings.TrimSpace(v)
	}
	if outPath == "" {
		v, err := prompt.AskText("Masukkan path file output hasil dekripsi: ", prompt.WithValidator(survey.Required))
		if err != nil {
			return err
		}
		outPath = strings.TrimSpace(v)
	}

	// Validate paths setelah user input (prevent path traversal)
	if err := cryptofile.ValidatePath(inPath); err != nil {
		return fmt.Errorf("invalid input path: %w", err)
	}
	if err := cryptofile.ValidatePath(outPath); err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	// Check if input file exists
	if _, err := os.Stat(inPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file input terenkripsi tidak ditemukan: %s", inPath)
		}
		return fmt.Errorf("gagal mengakses file input: %w", err)
	}

	// Check if output file exists and confirm overwrite
	if _, err := os.Stat(outPath); err == nil {
		confirm, err := prompt.AskConfirm(fmt.Sprintf("File '%s' sudah ada. Timpa?", outPath), false)
		if err != nil {
			return err
		}
		if !confirm {
			return fmt.Errorf("operasi dibatalkan oleh user")
		}
	}

	keyStr, _, err := ResolveKey(opts.Key, consts.ENV_ENCRYPTION_KEY, true)
	if err != nil {
		return err
	}

	err = DecryptFile(inPath, outPath, []byte(keyStr))
	audit.LogOperation(logger, audit.OpDecryptFile, inPath, err == nil, err)
	if err != nil {
		return err
	}

	print.PrintSubHeader("Hasil Dekripsi File")
	fmt.Printf("✓ File berhasil didekripsi\n")
	fmt.Printf("  Input:  %s\n", inPath)
	fmt.Printf("  Output: %s\n", outPath)
	return nil
}
