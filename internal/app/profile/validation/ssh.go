// File : internal/app/profile/validation/ssh.go
// Deskripsi : Validasi SSH tunnel (opsional)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/domain"
)

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
