// File : internal/shared/fsops/write.go
// Deskripsi : Helper functions untuk operasi write file dan directory
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 15 Januari 2026

package fsops

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile menulis data ke file di path yang diberikan.
// Security: tulis secara atomic (temp file + rename) dan permission 0600.
func WriteFile(filePath string, data []byte) error {
	if filePath == "" {
		return fmt.Errorf("filePath kosong")
	}

	dir := filepath.Dir(filePath)
	if dir == "" {
		return fmt.Errorf("direktori filePath tidak valid: %s", filePath)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}

	if err := tmp.Chmod(0o600); err != nil {
		cleanup()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Sync(); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}

	if err := os.Rename(tmpName, filePath); err != nil {
		// Windows compatibility: rename ke target existing bisa gagal.
		_ = os.Remove(filePath)
		if err2 := os.Rename(tmpName, filePath); err2 != nil {
			cleanup()
			return err
		}
	}

	return nil
}

// RemoveFile menghapus file di path yang diberikan.
func RemoveFile(path string) error {
	return os.Remove(path)
}

// CreateDirIfNotExist membuat direktori jika belum ada.
func CreateDirIfNotExist(dir string) error {
	if DirExists(dir) {
		return nil
	}
	// Hindari 0777. Permission final tetap dipengaruhi umask.
	return os.MkdirAll(dir, 0o755)
}
