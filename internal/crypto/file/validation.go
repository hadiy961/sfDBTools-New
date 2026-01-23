// File : internal/crypto/file/validation.go
// Deskripsi : Path validation helpers untuk security
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026
package file

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidatePath checks for path traversal attempts and suspicious patterns.
// Returns error if path contains dangerous patterns like ".." as a path component or null bytes.
func ValidatePath(path string) error {
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("invalid path: contains null byte")
	}

	// Normalisasi path untuk menghapus slash ganda, resolve . (current directory),
	// dan menormalkan separator ke OS-specific format.
	cleanPath := filepath.Clean(path)

	// Deteksi path traversal berbasis komponen, bukan sekadar substring.
	// Pisahkan path menggunakan pemisah '/' dan '\\' agar aman di berbagai OS.
	separator := func(r rune) bool {
		return r == '/' || r == '\\'
	}
	parts := strings.FieldsFunc(cleanPath, separator)
	for _, part := range parts {
		if part == ".." {
			return fmt.Errorf("invalid path: contains path traversal component '..': %s", path)
		}
	}
	return nil
}
