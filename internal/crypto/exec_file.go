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

// filePathConfig holds prompt messages for path input
type filePathConfig struct {
	InputPrompt       string
	OutputPrompt      string
	InputNotFoundMsg  string
	InputAccessErrMsg string
}

// validateAndPreparePaths handles common path validation logic for encrypt/decrypt operations.
// Returns validated input and output paths, or error if validation fails.
func validateAndPreparePaths(inPath, outPath string, cfg filePathConfig) (string, string, error) {
	// Trim and prompt for input path if empty
	inPath = strings.TrimSpace(inPath)
	if inPath == "" {
		v, err := prompt.AskText(cfg.InputPrompt, prompt.WithValidator(survey.Required))
		if err != nil {
			return "", "", err
		}
		inPath = strings.TrimSpace(v)
	}

	// Trim and prompt for output path if empty
	outPath = strings.TrimSpace(outPath)
	if outPath == "" {
		v, err := prompt.AskText(cfg.OutputPrompt, prompt.WithValidator(survey.Required))
		if err != nil {
			return "", "", err
		}
		outPath = strings.TrimSpace(v)
	}

	// Validate paths (prevent path traversal)
	if err := cryptofile.ValidatePath(inPath); err != nil {
		return "", "", fmt.Errorf("invalid input path: %w", err)
	}
	if err := cryptofile.ValidatePath(outPath); err != nil {
		return "", "", fmt.Errorf("invalid output path: %w", err)
	}

	// Check if input file exists
	if _, err := os.Stat(inPath); err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("%s: %s", cfg.InputNotFoundMsg, inPath)
		}
		return "", "", fmt.Errorf("%s: %w", cfg.InputAccessErrMsg, err)
	}

	// Check if output file exists and confirm overwrite
	if _, err := os.Stat(outPath); err == nil {
		confirm, err := prompt.AskConfirm(fmt.Sprintf("File '%s' sudah ada. Timpa?", outPath), false)
		if err != nil {
			return "", "", err
		}
		if !confirm {
			return "", "", fmt.Errorf("operasi dibatalkan oleh user")
		}
	}

	return inPath, outPath, nil
}

// ExecuteEncryptFile mengenkripsi file dengan AES-256-GCM.
func ExecuteEncryptFile(logger applog.Logger, opts EncryptFileOptions) error {
	cfg := filePathConfig{
		InputPrompt:       "Masukkan path file input: ",
		OutputPrompt:      "Masukkan path file output: ",
		InputNotFoundMsg:  "file input tidak ditemukan",
		InputAccessErrMsg: "gagal mengakses file input",
	}

	inPath, outPath, err := validateAndPreparePaths(opts.InputPath, opts.OutputPath, cfg)
	if err != nil {
		return err
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

	err = EncryptFile(inPath, outPath, keyBytes)
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
	cfg := filePathConfig{
		InputPrompt:       "Masukkan path file input terenkripsi: ",
		OutputPrompt:      "Masukkan path file output hasil dekripsi: ",
		InputNotFoundMsg:  "file input terenkripsi tidak ditemukan",
		InputAccessErrMsg: "gagal mengakses file input",
	}

	inPath, outPath, err := validateAndPreparePaths(opts.InputPath, opts.OutputPath, cfg)
	if err != nil {
		return err
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

	err = DecryptFile(inPath, outPath, keyBytes)
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
