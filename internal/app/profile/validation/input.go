// File : internal/app/profile/validation/input.go
// Deskripsi : Validasi input untuk wizard prompts
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package validation

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"sfdbtools/internal/app/profile/shared"
)

// ValidateNoLeadingTrailingSpace checks input tidak ada spasi di awal/akhir
func ValidateNoLeadingTrailingSpace(input string, fieldName string) error {
	if input != strings.TrimSpace(input) {
		if fieldName == "" {
			return shared.ErrInputHasLeadingTrailingSpace
		}
		return shared.ValidationError(fieldName, shared.ErrInputHasLeadingTrailingSpace)
	}
	return nil
}

// ValidateNoControlChars checks input tidak ada karakter kontrol
func ValidateNoControlChars(input string, fieldName string) error {
	for _, r := range input {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			if fieldName == "" {
				return shared.ErrInputHasControlChars
			}
			return shared.ValidationError(fieldName, shared.ErrInputHasControlChars)
		}
	}
	return nil
}

// ValidateNoSpaces checks input tidak ada spasi sama sekali
func ValidateNoSpaces(input string, fieldName string) error {
	if strings.Contains(input, " ") {
		if fieldName == "" {
			return shared.ErrInputHasSpaces
		}
		return shared.ValidationError(fieldName, shared.ErrInputHasSpaces)
	}
	return nil
}

// ValidateNotEmpty checks input tidak kosong
func ValidateNotEmpty(input string, fieldName string) error {
	if strings.TrimSpace(input) == "" {
		if fieldName == "" {
			return shared.ErrInputEmpty
		}
		return shared.ValidationError(fieldName, shared.ErrInputEmpty)
	}
	return nil
}

// ValidateIntInRange validates integer dalam range tertentu
func ValidateIntInRange(value, min, max int, allowZero bool, fieldName string) error {
	if allowZero && value == 0 {
		return nil
	}
	if value < min || value > max {
		return shared.InputRangeError(fieldName, min, max, allowZero)
	}
	return nil
}

// ValidateFileAccessible checks file accessible (bukan directory)
func ValidateFileAccessible(path string, fieldName string) error {
	p := strings.TrimSpace(path)
	if p == "" {
		return nil // Optional
	}
	info, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("%s tidak bisa diakses: %s", fieldName, p)
	}
	if info.IsDir() {
		return fmt.Errorf("%s tidak valid (path adalah direktori): %s", fieldName, p)
	}
	return nil
}

// ValidateConfigName validates nama konfigurasi profile
func ValidateConfigName(name string) error {
	if err := ValidateNotEmpty(name, "Nama Konfigurasi"); err != nil {
		return err
	}
	if err := ValidateNoLeadingTrailingSpace(name, "Nama Konfigurasi"); err != nil {
		return err
	}
	if err := ValidateNoControlChars(name, "Nama Konfigurasi"); err != nil {
		return err
	}
	return nil
}

// ValidateHost validates hostname/IP
func ValidateHost(host string) error {
	if err := ValidateNotEmpty(host, "Host"); err != nil {
		return err
	}
	if err := ValidateNoLeadingTrailingSpace(host, "Host"); err != nil {
		return err
	}
	if err := ValidateNoSpaces(host, "Host"); err != nil {
		return err
	}
	return nil
}

// ValidateUsername validates database/ssh username
func ValidateUsername(username string) error {
	if err := ValidateNotEmpty(username, "User"); err != nil {
		return err
	}
	if err := ValidateNoLeadingTrailingSpace(username, "User"); err != nil {
		return err
	}
	return nil
}
