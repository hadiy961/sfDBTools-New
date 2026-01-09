package fsops

import (
	"bufio"
	"os"
	"strings"
)

// WriteFile menulis data ke file di path yang diberikan dengan permission 0600.
func WriteFile(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, 0600)
}

// ReadFile membaca isi file dari path yang diberikan.
func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// RemoveFile menghapus file di path yang diberikan.
func RemoveFile(path string) error {
	return os.Remove(path)
}

// ReadLinesFromFile membaca semua baris dari file di path yang diberikan.
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

// CreateDirIfNotExist membuat direktori jika belum ada
func CreateDirIfNotExist(dir string) error {
	if DirExists(dir) {
		return nil
	}
	return os.MkdirAll(dir, os.ModePerm)
}
