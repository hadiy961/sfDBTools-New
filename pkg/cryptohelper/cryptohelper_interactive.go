// File : pkg/cryptohelper/cryptohelper_interactive.go
// Deskripsi : Helper functions untuk interactive mode input
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-11
// Last Modified : 2025-11-11

package cryptohelper

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// GetInteractiveInputBytes membaca input dari user secara interaktif (multi-line).
// User bisa paste text dan akhiri dengan Ctrl+D (Linux/Mac) atau Ctrl+Z (Windows).
//
// Parameter:
//   - prompt: pesan yang ditampilkan ke user
//
// Return:
//   - []byte: data yang diinput user
//   - error: error jika gagal membaca
func GetInteractiveInputBytes(prompt string) ([]byte, error) {
	fmt.Println(prompt)
	fmt.Println("(Paste teks Anda, lalu tekan Ctrl+D untuk selesai)")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	var lines []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// EOF reached (Ctrl+D)
			if line != "" {
				lines = append(lines, line)
			}
			break
		}
		lines = append(lines, line)
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("tidak ada input yang diberikan")
	}

	result := strings.Join(lines, "")
	return []byte(result), nil
}

// GetInteractiveInputString membaca input string dari user secara interaktif.
// Sama seperti GetInteractiveInputBytes tapi return string.
//
// Parameter:
//   - prompt: pesan yang ditampilkan ke user
//
// Return:
//   - string: data yang diinput user
//   - error: error jika gagal membaca
func GetInteractiveInputString(prompt string) (string, error) {
	data, err := GetInteractiveInputBytes(prompt)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetInputBytesOrInteractive mencoba membaca dari flag/pipe terlebih dahulu,
// jika tidak ada maka fallback ke interactive mode.
//
// Parameter:
//   - flagVal: nilai dari flag input
//   - prompt: pesan untuk interactive mode
//
// Return:
//   - []byte: data yang dibaca
//   - error: error jika gagal
func GetInputBytesOrInteractive(flagVal, prompt string) ([]byte, error) {
	// Coba dari flag
	if s := strings.TrimSpace(flagVal); s != "" {
		return []byte(s), nil
	}

	// Cek apakah ada pipe
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 { // piped
		reader := bufio.NewReader(os.Stdin)
		return bufio.NewReader(reader).ReadBytes('\x00') // Read all
	}

	// Fallback ke interactive mode
	return GetInteractiveInputBytes(prompt)
}

// GetInputStringOrInteractive mencoba membaca dari flag/pipe terlebih dahulu,
// jika tidak ada maka fallback ke interactive mode.
//
// Parameter:
//   - flagVal: nilai dari flag input
//   - prompt: pesan untuk interactive mode
//
// Return:
//   - string: data yang dibaca
//   - error: error jika gagal
func GetInputStringOrInteractive(flagVal, prompt string) (string, error) {
	// Coba dari flag
	if s := strings.TrimSpace(flagVal); s != "" {
		return s, nil
	}

	// Cek apakah ada pipe
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 { // piped
		reader := bufio.NewReader(os.Stdin)
		data, err := reader.ReadString('\x00') // Read all until null or EOF
		if err != nil && err.Error() != "EOF" {
			return "", err
		}
		return data, nil
	}

	// Fallback ke interactive mode
	return GetInteractiveInputString(prompt)
}

// GetFilePathOrInteractive membaca file path dari flag atau meminta input interaktif.
// Validasi bahwa file exists (untuk input file) atau direktori exists (untuk output file).
//
// Parameter:
//   - flagVal: nilai dari flag path
//   - prompt: pesan untuk interactive mode
//   - mustExist: true jika file harus sudah ada (input file), false untuk output file
//
// Return:
//   - string: file path yang valid
//   - error: error jika gagal atau file tidak valid
func GetFilePathOrInteractive(flagVal, prompt string, mustExist bool) (string, error) {
	// Jika ada flag value, gunakan itu
	if s := strings.TrimSpace(flagVal); s != "" {
		// Validasi file
		if mustExist {
			if _, err := os.Stat(s); os.IsNotExist(err) {
				return "", fmt.Errorf("file tidak ditemukan: %s", s)
			}
		}
		return s, nil
	}

	// Interactive mode
	fmt.Println(prompt)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Path: ")
		path, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("gagal membaca input: %w", err)
		}

		path = strings.TrimSpace(path)
		if path == "" {
			fmt.Println("✗ Path tidak boleh kosong. Coba lagi atau tekan Ctrl+C untuk batal.")
			continue
		}

		// Validasi
		if mustExist {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				fmt.Printf("✗ File tidak ditemukan: %s\n", path)
				fmt.Println("Coba lagi atau tekan Ctrl+C untuk batal.")
				continue
			}
		}

		return path, nil
	}
}
