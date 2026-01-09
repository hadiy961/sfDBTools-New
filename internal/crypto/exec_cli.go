// File : internal/crypto/exec_cli.go
// Deskripsi : Executor untuk perintah CLI crypto (thin orchestration)
// Author : Hadiyatna Muflihun
// Tanggal : 8 Januari 2026
// Last Modified : 9 Januari 2026

package crypto

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	cryptomodel "sfdbtools/internal/crypto/model"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/prompt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mattn/go-isatty"
)

func ExecuteBase64Encode(_ applog.Logger, opts cryptomodel.Base64EncodeOptions) error {
	var in io.Reader
	if strings.TrimSpace(opts.InputText) != "" {
		in = strings.NewReader(opts.InputText)
	} else {
		in = os.Stdin
	}

	out := io.Writer(os.Stdout)
	isStdout := true
	var f *os.File
	var err error
	if strings.TrimSpace(opts.OutputPath) != "" {
		f, err = os.Create(strings.TrimSpace(opts.OutputPath))
		if err != nil {
			return fmt.Errorf("gagal membuat file output: %w", err)
		}
		defer f.Close()
		out = f
		isStdout = false
	}

	enc := base64.NewEncoder(base64.StdEncoding, out)
	if _, err := io.Copy(enc, in); err != nil {
		_ = enc.Close()
		return fmt.Errorf("gagal encode base64: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("gagal finalize base64: %w", err)
	}

	if isStdout {
		_, _ = fmt.Fprintln(os.Stdout)
	}
	return nil
}

func ExecuteBase64Decode(_ applog.Logger, opts cryptomodel.Base64DecodeOptions) error {
	var in io.Reader
	if strings.TrimSpace(opts.InputData) != "" {
		in = strings.NewReader(opts.InputData)
	} else {
		in = os.Stdin
	}

	out := io.Writer(os.Stdout)
	var f *os.File
	var err error
	if strings.TrimSpace(opts.OutputPath) != "" {
		f, err = os.Create(strings.TrimSpace(opts.OutputPath))
		if err != nil {
			return fmt.Errorf("gagal membuat file output: %w", err)
		}
		defer f.Close()
		out = f
	}

	dec := base64.NewDecoder(base64.StdEncoding, in)
	if _, err := io.Copy(out, dec); err != nil {
		return fmt.Errorf("gagal decode base64: %w", err)
	}
	return nil
}

func ExecuteEncryptFile(_ applog.Logger, opts cryptomodel.EncryptFileOptions) error {
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

	keyStr, _, err := ResolveKey(opts.Key, consts.ENV_ENCRYPTION_KEY, true)
	if err != nil {
		return err
	}

	if err := EncryptFile(inPath, outPath, []byte(keyStr)); err != nil {
		return err
	}
	return nil
}

func ExecuteDecryptFile(_ applog.Logger, opts cryptomodel.DecryptFileOptions) error {
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

	keyStr, _, err := ResolveKey(opts.Key, consts.ENV_ENCRYPTION_KEY, true)
	if err != nil {
		return err
	}

	if err := DecryptFile(inPath, outPath, []byte(keyStr)); err != nil {
		return err
	}
	return nil
}

func ExecuteEncryptText(_ applog.Logger, opts cryptomodel.EncryptTextOptions) error {
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
			plaintext = []byte(v)
		} else {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("gagal membaca stdin: %w", err)
			}
			plaintext = b
		}
	}

	keyStr, _, err := ResolveKey(opts.Key, consts.ENV_ENCRYPTION_KEY, true)
	if err != nil {
		return err
	}

	cipherBytes, err := EncryptData(plaintext, []byte(keyStr))
	if err != nil {
		return err
	}

	outPath := strings.TrimSpace(opts.OutputPath)
	if outPath != "" {
		if err := os.WriteFile(outPath, cipherBytes, 0o600); err != nil {
			return fmt.Errorf("gagal menulis output: %w", err)
		}
		return nil
	}

	b64 := base64.StdEncoding.EncodeToString(cipherBytes)
	fmt.Println(b64)
	return nil
}

func ExecuteDecryptText(_ applog.Logger, opts cryptomodel.DecryptTextOptions) error {
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

	plain, err := DecryptData(cipherBytes, []byte(keyStr))
	if err != nil {
		return err
	}

	outPath := strings.TrimSpace(opts.OutputPath)
	if outPath != "" {
		if err := os.WriteFile(outPath, plain, 0o600); err != nil {
			return fmt.Errorf("gagal menulis output: %w", err)
		}
		return nil
	}

	_, err = os.Stdout.Write(plain)
	return err
}
