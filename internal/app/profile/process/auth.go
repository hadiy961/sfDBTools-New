// File : internal/app/profile/process/auth.go
// Deskripsi : Auth method builder untuk SSH tunnel
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 15 Januari 2026

package process

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func buildAuthMethods(opts SSHTunnelOptions) ([]ssh.AuthMethod, error) {
	methods := []ssh.AuthMethod{}

	if strings.TrimSpace(opts.IdentityFile) != "" {
		keyPath := opts.IdentityFile
		if !filepath.IsAbs(keyPath) {
			if wd, err := os.Getwd(); err == nil {
				keyPath = filepath.Join(wd, keyPath)
			}
		}
		keyPath = filepath.Clean(keyPath)
		b, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("gagal membaca identity file '%s': %w", keyPath, err)
		}
		signer, err := ssh.ParsePrivateKey(b)
		if err != nil {
			return nil, fmt.Errorf("gagal parse identity file '%s': %w", keyPath, err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	if strings.TrimSpace(opts.Password) != "" {
		methods = append(methods, ssh.Password(opts.Password))
	}

	// ssh-agent (jika tersedia)
	if sock := strings.TrimSpace(os.Getenv("SSH_AUTH_SOCK")); sock != "" {
		if conn, err := net.Dial("unix", sock); err == nil {
			ag := agent.NewClient(conn)
			if signers, serr := ag.Signers(); serr == nil && len(signers) > 0 {
				methods = append(methods, ssh.PublicKeys(signers...))
			}
			_ = conn.Close()
		}
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("metode autentikasi SSH tidak tersedia (isi ssh_password atau identity_file atau gunakan ssh-agent)")
	}
	return methods, nil
}
