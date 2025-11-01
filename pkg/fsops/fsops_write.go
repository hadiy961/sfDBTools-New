package fsops

import (
	"os"
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
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := []string{}
	currentLine := ""
	for _, b := range data {
		if b == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(b)
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines, nil
}

// EnsureDir ensures a directory exists
func EnsureDir(dir string) (string, error) {
	if dir == "" {
		return "", nil
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}
