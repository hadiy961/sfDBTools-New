package crypto

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"sfdbtools/internal/services/crypto/helpers"
	cryptokey "sfdbtools/internal/services/crypto/helpers"
	cryptomodel "sfdbtools/internal/services/crypto/model"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/ui/print"
	"sfdbtools/pkg/consts"
	"sfdbtools/pkg/encrypt"
)

// ExecuteEncryptText menangani logic encrypt text
func ExecuteEncryptText(logger applog.Logger, opts cryptomodel.EncryptTextOptions) error {
	quiet := helpers.SetupQuietMode(logger)

	// Password authentication
	MustValidatePassword()

	// Ambil input
	data, err := helpers.GetInput(opts.InputText, true, "Masukkan teks yang akan dienkripsi:")
	if err != nil {
		return fmt.Errorf("gagal membaca input: %v", err)
	}

	// Key
	key, _, err := cryptokey.ResolveEncryptionKey(opts.Key, consts.ENV_ENCRYPTION_KEY)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan encryption key: %v", err)
	}

	encBytes, err := encrypt.EncryptAES(data, []byte(key))
	if err != nil {
		return fmt.Errorf("gagal mengenkripsi teks: %v", err)
	}

	if strings.TrimSpace(opts.OutputPath) == "" {
		// Cetak base64 ke stdout
		if !quiet {
			print.PrintAppHeader("Encrypt Text")
			print.PrintSubHeader("Encrypted text output : ")
		}
		fmt.Println(base64.StdEncoding.EncodeToString(encBytes))
		if !quiet {
			print.PrintDashedSeparator()
		}
		return nil
	}

	// Tulis biner ke file
	if err := os.WriteFile(opts.OutputPath, encBytes, 0600); err != nil {
		return fmt.Errorf("gagal menulis file output: %v", err)
	}
	logger.Infof("✓ Teks terenkripsi ditulis ke file: %s", opts.OutputPath)
	return nil
}

// ExecuteDecryptText menangani logic decrypt text
func ExecuteDecryptText(logger applog.Logger, opts cryptomodel.DecryptTextOptions) error {
	quiet := helpers.SetupQuietMode(logger)

	// Password authentication
	MustValidatePassword()

	// Ambil input base64
	data, err := helpers.GetInput(opts.InputData, true, "Masukkan teks base64 terenkripsi:")
	if err != nil {
		return fmt.Errorf("gagal membaca input: %v", err)
	}
	b64 := string(data)

	encBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64))
	if err != nil {
		return fmt.Errorf("input bukan base64 yang valid: %v", err)
	}

	// Key
	key, _, err := cryptokey.ResolveEncryptionKey(opts.Key, consts.ENV_ENCRYPTION_KEY)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan encryption key: %v", err)
	}

	plain, err := encrypt.DecryptAES(encBytes, []byte(key))
	if err != nil {
		return fmt.Errorf("gagal mendekripsi teks: %v", err)
	}

	if strings.TrimSpace(opts.OutputPath) == "" {
		if !quiet {
			print.PrintAppHeader("Decrypt Text")
			print.PrintSubHeader("Decrypted text output : ")
		}
		fmt.Println(string(plain))
		if !quiet {
			print.PrintDashedSeparator()
		}
		return nil
	}

	if err := os.WriteFile(opts.OutputPath, plain, 0600); err != nil {
		return fmt.Errorf("gagal menulis file output: %v", err)
	}
	logger.Infof("✓ Teks didekripsi ditulis ke file: %s", opts.OutputPath)
	return nil
}
