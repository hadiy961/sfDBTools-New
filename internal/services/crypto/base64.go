package crypto

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"sfDBTools/internal/services/crypto/helpers"
	cryptomodel "sfDBTools/internal/services/crypto/model"
	applog "sfDBTools/internal/services/log"
	"sfDBTools/internal/ui/print"
)

// ExecuteBase64Encode menangani logic base64 encode
func ExecuteBase64Encode(logger applog.Logger, opts cryptomodel.Base64EncodeOptions) error {
	// Coba baca input dari flag atau pipe, fallback ke interactive
	b, err := helpers.GetInput(opts.InputText, true, "📝 Masukkan teks yang akan di-encode:")
	if err != nil {
		return fmt.Errorf("gagal membaca input: %v", err)
	}

	enc := base64.StdEncoding.EncodeToString(b)
	print.PrintAppHeader("Base64 Encode")
	if strings.TrimSpace(opts.OutputPath) == "" {
		print.PrintSubHeader("Output text")
		fmt.Println(enc)
		print.PrintDashedSeparator()
		return nil
	}
	if err := os.WriteFile(opts.OutputPath, []byte(enc), 0644); err != nil {
		return fmt.Errorf("gagal menulis file: %v", err)
	}
	logger.Infof("✓ Base64 tersimpan: %s", opts.OutputPath)
	return nil
}

// ExecuteBase64Decode menangani logic base64 decode
func ExecuteBase64Decode(logger applog.Logger, opts cryptomodel.Base64DecodeOptions) error {
	// Coba baca input dari flag atau pipe, fallback ke interactive
	data, err := helpers.GetInput(opts.InputData, true, "📝 Masukkan base64 yang akan di-decode:")
	if err != nil {
		return fmt.Errorf("gagal membaca input: %v", err)
	}
	in := string(data)

	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(in))
	if err != nil {
		return fmt.Errorf("input bukan base64 yang valid: %v", err)
	}
	print.PrintAppHeader("Base64 Decode")
	print.PrintSubHeader("Output Text : ")
	if strings.TrimSpace(opts.OutputPath) == "" {
		fmt.Println(string(b))
		print.PrintDashedSeparator()
		return nil
	}
	if err := os.WriteFile(opts.OutputPath, b, 0644); err != nil {
		return fmt.Errorf("gagal menulis file: %v", err)
	}
	logger.Infof("✓ Decode tersimpan: %s", opts.OutputPath)
	return nil
}
