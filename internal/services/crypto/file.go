package crypto

import (
	"fmt"
	"io"
	"os"

	"sfdbtools/internal/services/crypto/helpers"
	cryptomodel "sfdbtools/internal/services/crypto/model"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/ui/print"
	"sfdbtools/pkg/consts"
	"sfdbtools/pkg/encrypt"
	"sfdbtools/pkg/helper"
)

// ExecuteEncryptFile menangani logic encrypt file
func ExecuteEncryptFile(logger applog.Logger, opts cryptomodel.EncryptFileOptions) error {
	quiet := helpers.SetupQuietMode(logger)

	// Password authentication
	MustValidatePassword()

	if !quiet {
		print.PrintAppHeader("Encrypt File")
	}

	// Interactive mode untuk file paths jika tidak ada flag
	inputPath, err := helpers.GetFilePathOrInteractive(opts.InputPath, "📂 Masukkan path file yang akan dienkripsi:", true)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan input file: %v", err)
	}

	outputPath, err := helpers.GetFilePathOrInteractive(opts.OutputPath, "💾 Masukkan path file output terenkripsi:", false)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan output file: %v", err)
	}

	// Dapatkan key dari flag/env/prompt
	key, _, err := helper.ResolveEncryptionKey(opts.Key, consts.ENV_ENCRYPTION_KEY)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan encryption key: %v", err)
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
				return fmt.Errorf("gagal membuka file input: %v", err)
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
				return fmt.Errorf("gagal membuat file output: %v", err)
			}
			defer f.Close()
			writer = f
		}

		ew, err := encrypt.NewEncryptingWriter(writer, []byte(key))
		if err != nil {
			return fmt.Errorf("gagal membuat encrypting writer: %v", err)
		}
		if _, err := io.Copy(ew, reader); err != nil {
			return fmt.Errorf("gagal mengenkripsi data: %v", err)
		}
		if err := ew.Close(); err != nil {
			return fmt.Errorf("gagal menutup writer: %v", err)
		}
	} else {
		if err := encrypt.EncryptFile(inputPath, outputPath, []byte(key)); err != nil {
			return fmt.Errorf("gagal mengenkripsi file: %v", err)
		}
		logger.Infof("✓ File terenkripsi: %s → %s", inputPath, outputPath)
	}
	return nil
}

// ExecuteDecryptFile menangani logic decrypt file
func ExecuteDecryptFile(logger applog.Logger, opts cryptomodel.DecryptFileOptions) error {
	quiet := helpers.SetupQuietMode(logger)

	// Password authentication
	MustValidatePassword()

	if !quiet {
		print.PrintAppHeader("Decrypt File")
	}

	// Interactive mode untuk file paths jika tidak ada flag
	inputPath, err := helpers.GetFilePathOrInteractive(opts.InputPath, "📂 Masukkan path file terenkripsi yang akan didekripsi:", true)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan input file: %v", err)
	}

	outputPath, err := helpers.GetFilePathOrInteractive(opts.OutputPath, "💾 Masukkan path file output hasil dekripsi:", false)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan output file: %v", err)
	}

	// Dapatkan key dari flag/env/prompt
	key, _, err := helper.ResolveEncryptionKey(opts.Key, consts.ENV_ENCRYPTION_KEY)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan encryption key: %v", err)
	}

	// Support pipeline via '-' for stdin/stdout
	if inputPath == "-" || outputPath == "-" {
		// Read all from input (stdin or file), decrypt, then write to stdout/file
		var data []byte
		if inputPath == "-" {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("gagal membaca stdin: %v", err)
			}
			data = b
		} else {
			b, err := os.ReadFile(inputPath)
			if err != nil {
				return fmt.Errorf("gagal membaca file input: %v", err)
			}
			data = b
		}

		plain, err := encrypt.DecryptAES(data, []byte(key))
		if err != nil {
			return fmt.Errorf("gagal mendekripsi data: %v", err)
		}

		if outputPath == "-" {
			if _, err := os.Stdout.Write(plain); err != nil {
				return fmt.Errorf("gagal menulis stdout: %v", err)
			}
		} else {
			if err := os.WriteFile(outputPath, plain, 0600); err != nil {
				return fmt.Errorf("gagal menulis file output: %v", err)
			}
			logger.Infof("✓ File didekripsi: %s → %s", inputPath, outputPath)
		}
	} else {
		if err := encrypt.DecryptFile(inputPath, outputPath, []byte(key)); err != nil {
			return fmt.Errorf("gagal mendekripsi file: %v", err)
		}
		logger.Infof("✓ File didekripsi: %s → %s", inputPath, outputPath)
	}
	return nil
}
