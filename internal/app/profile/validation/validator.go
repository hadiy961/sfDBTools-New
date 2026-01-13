// File : internal/app/profile/validation/validator.go
// Deskripsi : Centralized validation logic untuk profile module (P1 improvement)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/domain"
)

// =============================================================================
// Profile Info Validation
// =============================================================================

// ValidateProfileInfo melakukan validasi komprehensif terhadap ProfileInfo
func ValidateProfileInfo(p *domain.ProfileInfo) error {
	if p == nil {
		return shared.ErrProfileNil
	}
	if strings.TrimSpace(p.Name) == "" {
		return shared.ErrProfileNameEmpty
	}
	if err := ValidateDBInfo(&p.DBInfo); err != nil {
		return fmt.Errorf("validasi db info gagal: %w", err)
	}
	if p.SSHTunnel.Enabled {
		if strings.TrimSpace(p.SSHTunnel.Host) == "" {
			return shared.ErrSSHTunnelHostEmpty
		}
	}
	return nil
}

// ValidateDBInfo melakukan validasi terhadap DBInfo
func ValidateDBInfo(db *domain.DBInfo) error {
	if db == nil {
		return shared.ErrDBInfoNil
	}
	if strings.TrimSpace(db.Host) == "" {
		return shared.ErrDBHostEmpty
	}
	if db.Port <= 0 || db.Port > 65535 {
		return shared.DBPortInvalidError(db.Port)
	}
	if strings.TrimSpace(db.User) == "" {
		return shared.ErrDBUserEmpty
	}
	// Password bisa kosong untuk beberapa auth method
	return nil
}

// =============================================================================
// SSH Validation
// =============================================================================

// ValidateSSHTunnel validates SSH tunnel configuration
func ValidateSSHTunnel(ssh *domain.SSHTunnelConfig) error {
	if ssh == nil {
		return nil // SSH optional
	}
	if !ssh.Enabled {
		return nil
	}
	if strings.TrimSpace(ssh.Host) == "" {
		return shared.ErrSSHHostEmpty
	}
	if ssh.Port <= 0 || ssh.Port > 65535 {
		return shared.SSHPortInvalidError(ssh.Port)
	}
	// Validate identity file if provided
	if ssh.IdentityFile != "" {
		if err := ValidateSSHIdentityFile(ssh.IdentityFile); err != nil {
			return err
		}
	}
	return nil
}

// ValidateSSHIdentityFile validates SSH identity file accessibility
func ValidateSSHIdentityFile(path string) error {
	p := strings.TrimSpace(path)
	if p == "" {
		return nil // Optional
	}
	// Expand home dir
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			p = filepath.Join(home, p[2:])
		}
	}
	info, err := os.Stat(p)
	if err != nil {
		return shared.SSHIdentityFileError(p, fmt.Sprintf("tidak bisa diakses: %v", err))
	}
	if info.IsDir() {
		return shared.SSHIdentityFileError(p, "path adalah direktori")
	}
	// Try to read (check permission)
	f, err := os.Open(p)
	if err != nil {
		return shared.SSHIdentityFileError(p, fmt.Sprintf("tidak bisa dibaca: %v", err))
	}
	f.Close()
	return nil
}

// =============================================================================
// Input Validation (untuk wizard prompts)
// =============================================================================

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

// =============================================================================
// Combined Validators (chains multiple validations)
// =============================================================================

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
