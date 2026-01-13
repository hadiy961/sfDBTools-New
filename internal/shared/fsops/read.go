// File : internal/shared/fsops/read.go
// Deskripsi : Helper functions untuk operasi read file dan directory
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package fsops

import (
	"bufio"
	"os"
	"strings"
)

// ReadFile membaca isi file dari path yang diberikan.
func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ReadDirFiles membaca nama-nama file dalam direktori yang diberikan.
// Return hanya file (bukan directory), tidak recursive.
func ReadDirFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}
	return files, nil
}

// ReadLinesFromFile membaca semua baris dari file di path yang diberikan.
// Normalisasi CRLF (Windows) ke LF otomatis.
func ReadLinesFromFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Default token size 64K kadang terlalu kecil (mis. line panjang). Naikkan limit.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lines := make([]string, 0, 64)
	for scanner.Scan() {
		// Normalisasi CRLF (Windows) agar downstream list clean.
		line := strings.TrimSuffix(scanner.Text(), "\r")
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
