// File : internal/shared/fsops/write.go
// Deskripsi : Helper functions untuk operasi write file dan directory
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package fsops

import "os"

// WriteFile menulis data ke file di path yang diberikan.
// Permission 0600 (read/write owner only) untuk keamanan file sensitif.
func WriteFile(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, 0600)
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
	return os.MkdirAll(dir, os.ModePerm)
}
