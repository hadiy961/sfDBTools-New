// File : internal/app/profile/helpers/connect_preflight.go
// Deskripsi : Preflight validation untuk koneksi DB dan SSH tunnel sebelum mencoba dial
// Author : Hadiyatna Muflihun
// Tanggal : 9 Januari 2026
// Last Modified : 9 Januari 2026

package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfdbtools/internal/domain"
)

func ValidateConnectPreflight(profile *domain.ProfileInfo) error {
	if profile == nil {
		return fmt.Errorf("profile tidak boleh nil")
	}

	host := strings.TrimSpace(profile.DBInfo.Host)
	user := strings.TrimSpace(profile.DBInfo.User)
	port := profile.DBInfo.Port
	if host == "" {
		return fmt.Errorf("host database kosong")
	}
	if port <= 0 || port > 65535 {
		return fmt.Errorf("port database tidak valid: %d", port)
	}
	if user == "" {
		return fmt.Errorf("user database kosong")
	}

	if !profile.SSHTunnel.Enabled {
		return nil
	}

	sshHost := strings.TrimSpace(profile.SSHTunnel.Host)
	sshPort := profile.SSHTunnel.Port
	localPort := profile.SSHTunnel.LocalPort
	identity := strings.TrimSpace(profile.SSHTunnel.IdentityFile)

	if sshHost == "" {
		return fmt.Errorf("ssh tunnel aktif tetapi ssh host kosong")
	}
	if sshPort == 0 {
		sshPort = 22
	}
	if sshPort <= 0 || sshPort > 65535 {
		return fmt.Errorf("ssh port tidak valid: %d", sshPort)
	}
	if localPort < 0 || localPort > 65535 {
		return fmt.Errorf("ssh local port tidak valid: %d", localPort)
	}

	if identity != "" {
		p := identity
		if !filepath.IsAbs(p) {
			if wd, err := os.Getwd(); err == nil {
				p = filepath.Join(wd, p)
			}
		}
		p = filepath.Clean(p)
		fi, err := os.Stat(p)
		if err != nil {
			return fmt.Errorf("identity file SSH tidak bisa diakses: %s (%v)", p, err)
		}
		if fi.IsDir() {
			return fmt.Errorf("identity file SSH tidak valid (path adalah direktori): %s", p)
		}
		f, err := os.Open(p)
		if err != nil {
			return fmt.Errorf("identity file SSH tidak bisa dibaca: %s (%v)", p, err)
		}
		_ = f.Close()
	}

	return nil
}
