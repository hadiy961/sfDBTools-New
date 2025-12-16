package cryptohelper

// File : pkg/cryptohelper/cryptohelper_input.go
// Deskripsi : Helper functions untuk membaca input dari flag atau stdin
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-11
// Last Modified : 2025-11-11

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// GetInput membaca input dari flag, pipe stdin, atau fallback ke interactive prompt.
// Parameter:
//   - flagVal: nilai dari flag input (kosong jika tidak ada)
//   - allowInteractive: jika true, fallback ke interactive prompt
//   - prompt: pesan untuk interactive mode (hanya jika allowInteractive=true)
//
// Return:
//   - []byte: data yang dibaca
//   - error: error jika tidak ada input atau gagal membaca
func GetInput(flagVal string, allowInteractive bool, prompt string) ([]byte, error) {
	if s := strings.TrimSpace(flagVal); s != "" {
		return []byte(s), nil
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return io.ReadAll(bufio.NewReader(os.Stdin))
	}

	if allowInteractive {
		return GetInteractiveInputBytes(prompt)
	}

	return nil, fmt.Errorf("tidak ada input: berikan flag input atau pipe melalui stdin")
}
