// File : internal/crypto/exec_base64.go
// Deskripsi : Executor untuk base64 encode/decode CLI commands
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
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mattn/go-isatty"
)

// ExecuteBase64Encode melakukan encoding base64 dari text/stdin ke stdout/file.
func ExecuteBase64Encode(logger applog.Logger, opts Base64EncodeOptions) error {
	var in io.Reader
	if strings.TrimSpace(opts.InputText) != "" {
		in = strings.NewReader(opts.InputText)
	} else {
		// Jika stdin adalah TTY, prompt user untuk input
		if isatty.IsTerminal(os.Stdin.Fd()) {
			v, err := prompt.AskText("Masukkan teks untuk di-encode (base64): ", prompt.WithValidator(survey.Required))
			if err != nil {
				return err
			}
			in = strings.NewReader(v)
		} else {
			in = os.Stdin
		}
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

	if isStdout {
		print.PrintSubHeader("Hasil Base64 Encode")
	}

	enc := base64.NewEncoder(base64.StdEncoding, out)
	var encErr error
	if _, encErr = io.Copy(enc, in); encErr != nil {
		_ = enc.Close()
		audit.LogOperation(logger, audit.OpBase64Enc, "<stdin>", false, encErr)
		return fmt.Errorf("gagal encode base64: %w", encErr)
	}
	if encErr = enc.Close(); encErr != nil {
		audit.LogOperation(logger, audit.OpBase64Enc, "<stdin>", false, encErr)
		return fmt.Errorf("gagal finalize base64: %w", encErr)
	}
	audit.LogOperation(logger, audit.OpBase64Enc, "<stdin>", true, nil)

	// Always output newline untuk consistency (scripting-friendly)
	if isStdout {
		_, _ = fmt.Fprintln(os.Stdout)
	} else {
		// Untuk file output juga, tambahkan newline untuk consistency
		_, _ = f.Write([]byte("\n"))
	}
	return nil
}

// ExecuteBase64Decode melakukan decoding base64 dari text/stdin ke stdout/file.
func ExecuteBase64Decode(logger applog.Logger, opts Base64DecodeOptions) error {
	var in io.Reader
	if strings.TrimSpace(opts.InputData) != "" {
		in = strings.NewReader(opts.InputData)
	} else {
		// Jika stdin adalah TTY, prompt user untuk input
		if isatty.IsTerminal(os.Stdin.Fd()) {
			v, err := prompt.AskText("Masukkan teks base64 untuk di-decode: ", prompt.WithValidator(survey.Required))
			if err != nil {
				return err
			}
			in = strings.NewReader(v)
		} else {
			in = os.Stdin
		}
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

	if isStdout {
		print.PrintSubHeader("Hasil Base64 Decode")
	}

	dec := base64.NewDecoder(base64.StdEncoding, in)
	var decErr error
	if _, decErr = io.Copy(out, dec); decErr != nil {
		audit.LogOperation(logger, audit.OpBase64Dec, "<stdin>", false, decErr)
		return fmt.Errorf("gagal decode base64: %w\n\nPossible causes:\n  - Input bukan base64 yang valid\n  - Input corrupted atau incomplete\n\nHint: Pastikan input adalah base64 encoding yang benar", decErr)
	}

	// Catatan: Tidak menambahkan newline di akhir output.
	// Decoder base64 harus mempertahankan data asli byte-for-byte, termasuk untuk file biner.
	audit.LogOperation(logger, audit.OpBase64Dec, "<stdin>", true, nil)
	return nil
}
