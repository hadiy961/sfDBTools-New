// File : internal/services/crypto/helpers/interactive.go
// Deskripsi : Helper functions untuk interactive mode input
// Author : Hadiyatna Muflihun
// Tanggal : 11 November 2025
// Last Modified : 5 Januari 2026
package helpers

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"sfdbtools/pkg/fsops"
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

// GetFilePathOrInteractive membaca file path dari flag atau meminta input interaktif.
// Parameter:
//   - flagVal: nilai dari flag path
//   - prompt: pesan untuk interactive mode
//   - mustExist: true jika file harus sudah ada (input file), false untuk output file
//
// Return:
//   - string: file path yang valid
//   - error: error jika gagal atau file tidak valid
func GetFilePathOrInteractive(flagVal, prompt string, mustExist bool) (string, error) {
	if s := strings.TrimSpace(flagVal); s != "" {
		if mustExist && !fsops.FileExists(s) {
			return "", fmt.Errorf("file tidak ditemukan: %s", s)
		}
		return s, nil
	}

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

		if mustExist && !fsops.FileExists(path) {
			fmt.Printf("✗ File tidak ditemukan: %s\n", path)
			fmt.Println("Coba lagi atau tekan Ctrl+C untuk batal.")
			continue
		}

		return path, nil
	}
}
