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
// Returns error if path contains dangerous patterns like ".." or null bytes.
func ValidatePath(path string) error {
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("invalid path: contains null byte")
	}
	// Check for path traversal
	clean := filepath.Clean(path)
	if strings.Contains(clean, "..") {
		return fmt.Errorf("invalid path: contains path traversal (..): %s", path)
	}
	return nil
}
