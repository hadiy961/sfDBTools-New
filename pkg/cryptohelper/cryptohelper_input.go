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

// GetInputBytes membaca byte data dari flag string atau STDIN jika tersedia.
// Digunakan untuk membaca data binary atau teks untuk enkripsi/encode.
//
// Parameter:
//   - flagVal: nilai dari flag input (kosong jika tidak ada)
//
// Return:
//   - []byte: data yang dibaca
//   - error: error jika tidak ada input atau gagal membaca
func GetInputBytes(flagVal string) ([]byte, error) {
	if s := strings.TrimSpace(flagVal); s != "" {
		return []byte(s), nil
	}

	// Cek apakah ada data dari stdin (pipe)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 { // piped
		reader := bufio.NewReader(os.Stdin)
		return io.ReadAll(reader)
	}

	return nil, fmt.Errorf("tidak ada input: berikan flag input atau pipe melalui stdin")
}

// GetInputString membaca string dari flag atau STDIN jika tersedia.
// Digunakan untuk membaca teks seperti base64 atau data terenkripsi.
//
// Parameter:
//   - flagVal: nilai dari flag input (kosong jika tidak ada)
//
// Return:
//   - string: data yang dibaca
//   - error: error jika tidak ada input atau gagal membaca
func GetInputString(flagVal string) (string, error) {
	if s := strings.TrimSpace(flagVal); s != "" {
		return s, nil
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		reader := bufio.NewReader(os.Stdin)
		b, err := io.ReadAll(reader)
		return string(b), err
	}

	return "", fmt.Errorf("tidak ada input: berikan flag input atau pipe melalui stdin")
}
