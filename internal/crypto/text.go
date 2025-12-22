package crypto

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/cryptoauth"
	"sfDBTools/pkg/cryptohelper"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
)

// ExecuteEncryptText menangani logic encrypt text
func ExecuteEncryptText(logger applog.Logger, opts types.EncryptTextOptions) error {
	quiet := cryptohelper.SetupQuietMode(logger)

	// Password authentication
	cryptoauth.MustValidatePassword()

	// Ambil input
	data, err := cryptohelper.GetInput(opts.InputText, true, "Masukkan teks yang akan dienkripsi:")
	if err != nil {
		return fmt.Errorf("gagal membaca input: %v", err)
	}

	// Key
	key, _, err := helper.ResolveEncryptionKey(opts.Key, consts.ENV_ENCRYPTION_KEY)
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
			ui.Headers("Encrypt Text")
			ui.PrintSubHeader("Encrypted text output : ")
		}
		fmt.Println(base64.StdEncoding.EncodeToString(encBytes))
		if !quiet {
			ui.PrintDashedSeparator()
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
func ExecuteDecryptText(logger applog.Logger, opts types.DecryptTextOptions) error {
	quiet := cryptohelper.SetupQuietMode(logger)

	// Password authentication
	cryptoauth.MustValidatePassword()

	// Ambil input base64
	data, err := cryptohelper.GetInput(opts.InputData, true, "Masukkan teks base64 terenkripsi:")
	if err != nil {
		return fmt.Errorf("gagal membaca input: %v", err)
	}
	b64 := string(data)

	encBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64))
	if err != nil {
		return fmt.Errorf("input bukan base64 yang valid: %v", err)
	}

	// Key
	key, _, err := helper.ResolveEncryptionKey(opts.Key, consts.ENV_ENCRYPTION_KEY)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan encryption key: %v", err)
	}

	plain, err := encrypt.DecryptAES(encBytes, []byte(key))
	if err != nil {
		return fmt.Errorf("gagal mendekripsi teks: %v", err)
	}

	if strings.TrimSpace(opts.OutputPath) == "" {
		if !quiet {
			ui.Headers("Decrypt Text")
			ui.PrintSubHeader("Decrypted text output : ")
		}
		fmt.Println(string(plain))
		if !quiet {
			ui.PrintDashedSeparator()
		}
		return nil
	}

	if err := os.WriteFile(opts.OutputPath, plain, 0600); err != nil {
		return fmt.Errorf("gagal menulis file output: %v", err)
	}
	logger.Infof("✓ Teks didekripsi ditulis ke file: %s", opts.OutputPath)
	return nil
}
