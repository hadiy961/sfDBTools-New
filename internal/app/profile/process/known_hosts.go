// File : internal/app/profile/process/known_hosts.go
// Deskripsi : Known hosts handling untuk SSH tunnel
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 15 Januari 2026

package process

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sfdbtools/internal/shared/consts"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const defaultSfDBToolsKnownHostsPath = "/etc/sfDBTools/known_hosts"

const fallbackUserKnownHostsPath = ".config/sfdbtools/known_hosts"

func ensureFile(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path kosong")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	return f.Close()
}

func selectKnownHostsPath() (string, error) {
	// 1) Prefer /etc/sfDBTools/known_hosts
	if err := ensureFile(defaultSfDBToolsKnownHostsPath); err == nil {
		return defaultSfDBToolsKnownHostsPath, nil
	}

	// 2) Fallback per-user: ~/.config/sfdbtools/known_hosts
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "", fmt.Errorf("tidak bisa menentukan home dir untuk fallback known_hosts")
	}
	p := filepath.Join(home, fallbackUserKnownHostsPath)
	if err := ensureFile(p); err != nil {
		return "", err
	}
	return p, nil
}

func hostPatterns(host string, port int) []string {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil
	}
	if port == 0 || port == 22 {
		return []string{host}
	}
	return []string{fmt.Sprintf("[%s]:%d", host, port)}
}

func appendKnownHost(path string, host string, port int, key ssh.PublicKey) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("known_hosts path kosong")
	}
	patterns := hostPatterns(host, port)
	if len(patterns) == 0 {
		return fmt.Errorf("host kosong")
	}
	line := knownhosts.Line(patterns, key)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(line + "\n"); err != nil {
		return err
	}
	return nil
}

func buildHostKeyCallbackFor(host string, port int) (ssh.HostKeyCallback, string, error) {
	knownHostsPath, err := selectKnownHostsPath()
	if err != nil {
		// Security: jangan silently insecure. Izinkan bypass hanya jika user opt-in via env.
		if strings.TrimSpace(os.Getenv(consts.ENV_SSH_INSECURE_IGNORE_HOSTKEY)) == "1" {
			return ssh.InsecureIgnoreHostKey(), "", nil
		}
		return nil, "", fmt.Errorf(
			"tidak bisa menentukan known_hosts path (verifikasi host key SSH tidak bisa dilakukan). "+
				"Perbaiki permission/home dir atau set %s=1 untuk bypass (TIDAK AMAN): %w",
			consts.ENV_SSH_INSECURE_IGNORE_HOSTKEY,
			err,
		)
	}

	cb, kerr := knownhosts.New(knownHostsPath)
	if kerr != nil {
		return nil, knownHostsPath, fmt.Errorf("gagal membaca known_hosts (%s): %w", knownHostsPath, kerr)
	}

	// Auto-trust seperti client GUI: jika host key belum ada / mismatch,
	// tambahkan key yang sedang dipresentasikan ke file known_hosts sfdbtools.
	wrapped := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if err := cb(hostname, remote, key); err == nil {
			return nil
		} else {
			var keyErr *knownhosts.KeyError
			if errors.As(err, &keyErr) {
				// Jangan sentuh ~/.ssh/known_hosts; tulis ke file khusus sfdbtools.
				if werr := appendKnownHost(knownHostsPath, host, port, key); werr == nil {
					return nil
				}
			}
			return err
		}
	}

	return wrapped, knownHostsPath, nil
}
