package fsops

import (
	"fmt"
	"os"
	"path/filepath"
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
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// CreateBackupDirs membuat direktori base dan (opsional) subdirektori berdasarkan konfigurasi.
// Mengembalikan path final tempat backup akan disimpan.
func CreateBackupDirs(baseDir string, createSubdirs bool, structurePattern, client string) (bool, error) {
	if baseDir == "" {
		return false, fmt.Errorf("direktori dasar kosong")
	}

	// Pastikan direktori dasar ada
	dir, err := CheckDirExists(baseDir)
	if err != nil {
		return false, fmt.Errorf("gagal memastikan direktori dasar: %w", err)
	}
	if !dir {
		// fmt.Println("Membuat direktori dasar:", baseDir)
		if err := CreateDirIfNotExist(baseDir); err != nil {
			return false, fmt.Errorf("gagal membuat direktori dasar: %w", err)
		}
	}

	finalDir := baseDir
	if createSubdirs {
		subdir, err := BuildSubdirPath(structurePattern, client)
		if err != nil {
			return false, fmt.Errorf("gagal membangun path subdirektori: %w", err)
		}
		finalDir = filepath.Join(baseDir, subdir)
		if err := CreateDirIfNotExist(finalDir); err != nil {
			return false, fmt.Errorf("gagal membuat path subdirektori: %w", err)
		}
	}
	return true, nil
}

// CreateDirIfNotExist membuat direktori jika belum ada
func CreateDirIfNotExist(dir string) error {
	// Cek apakah direktori sudah ada
	exists, err := CheckDirExists(dir)
	if err != nil {
		return err
	}
	if !exists {
		// Buat direktori beserta parent-nya
		if err := CreateDir(dir); err != nil {
			return err
		}
	}
	return nil
}

// CreateDir membuat direktori beserta parent-nya jika belum ada
func CreateDir(dir string) error {
	return os.MkdirAll(dir, os.ModePerm)
}
