// File : internal/app/profile/process/ssh_tunnel.go
// Deskripsi : SSH tunnel native Go (port forwarding)
// Author : Hadiyatna Muflihun
// Tanggal : 2 Januari 2026
// Last Modified : 9 Januari 2026

package process

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

const defaultSfDBToolsKnownHostsPath = "/etc/sfDBTools/known_hosts"

const fallbackUserKnownHostsPath = ".config/sfdbtools/known_hosts"

type SSHTunnelOptions struct {
	SSHHost        string
	SSHPort        int
	SSHUser        string
	Password       string
	IdentityFile   string
	LocalPort      int
	RemoteHost     string
	RemotePort     int
	ConnectTimeout time.Duration
	ServerAlive    time.Duration
	ExitOnFailure  bool
	BatchMode      bool
}

type SSHTunnel struct {
	LocalPort int
	listener  net.Listener
	sshClient *ssh.Client
	cancel    context.CancelFunc
	once      sync.Once
}

func pickLocalPort(requested int) (int, error) {
	if requested > 0 {
		return requested, nil
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	addr := ln.Addr().String()
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, err
	}
	p, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, err
	}
	return p, nil
}

func (o SSHTunnelOptions) validate() error {
	o.SSHHost = strings.TrimSpace(o.SSHHost)
	o.SSHUser = strings.TrimSpace(o.SSHUser)
	o.Password = strings.TrimSpace(o.Password)
	o.IdentityFile = strings.TrimSpace(o.IdentityFile)
	o.RemoteHost = strings.TrimSpace(o.RemoteHost)
	if o.SSHHost == "" {
		return fmt.Errorf("ssh host kosong")
	}
	if o.RemoteHost == "" {
		return fmt.Errorf("remote host kosong")
	}
	if o.RemotePort == 0 {
		return fmt.Errorf("remote port kosong")
	}
	if o.SSHPort == 0 {
		o.SSHPort = 22
	}
	return nil
}

func defaultKnownHostsPaths() []string {
	// Prioritaskan known_hosts khusus sfdbtools agar tidak konflik dengan ~/.ssh/known_hosts.
	// Jika file ini ada, kita akan pakai secara eksklusif.
	paths := []string{defaultSfDBToolsKnownHostsPath}
	if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
		paths = append(paths, filepath.Join(home, ".ssh", "known_hosts"))
	}
	paths = append(paths, "/etc/ssh/ssh_known_hosts")
	return paths
}

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
		// Fallback agar tetap bisa jalan (terutama di environment yang tidak punya permission menulis file).
		// Ini mengurangi keamanan verifikasi host key.
		return ssh.InsecureIgnoreHostKey(), "", nil
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

// StartSSHTunnel memulai SSH tunnel (local port forwarding) menggunakan native Go.
func StartSSHTunnel(ctx context.Context, opts SSHTunnelOptions) (*SSHTunnel, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	localPort, err := pickLocalPort(opts.LocalPort)
	if err != nil {
		return nil, fmt.Errorf("gagal menentukan local port: %w", err)
	}

	sshPort := opts.SSHPort
	if sshPort == 0 {
		sshPort = 22
	}

	connectTimeout := opts.ConnectTimeout
	if connectTimeout == 0 {
		connectTimeout = 10 * time.Second
	}

	cb, usedKnownHosts, err := buildHostKeyCallbackFor(opts.SSHHost, sshPort)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat host key callback: %w", err)
	}

	authMethods, err := buildAuthMethods(opts)
	if err != nil {
		return nil, err
	}

	user := strings.TrimSpace(opts.SSHUser)
	if user == "" {
		user = os.Getenv("USER")
	}
	if strings.TrimSpace(user) == "" {
		user = "root"
	}

	sshCfg := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: cb,
		Timeout:         connectTimeout,
	}

	sshAddr := net.JoinHostPort(opts.SSHHost, strconv.Itoa(sshPort))

	ctx2, cancel := context.WithCancel(ctx)
	// Dial TCP dulu agar honor connectTimeout.
	tcpConn, err := net.DialTimeout("tcp", sshAddr, connectTimeout)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("gagal konek TCP ke ssh %s: %w", sshAddr, err)
	}

	cc, chans, reqs, err := ssh.NewClientConn(tcpConn, sshAddr, sshCfg)
	if err != nil {
		_ = tcpConn.Close()
		cancel()
		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) {
			kh := usedKnownHosts
			if strings.TrimSpace(kh) == "" {
				kh = defaultSfDBToolsKnownHostsPath
			}
			return nil, fmt.Errorf(
				"ssh handshake gagal ke %s: known_hosts key mismatch. sfdbtools akan memakai file: %s. Pastikan host yang dituju benar lalu update host key di file tsb: %w",
				sshAddr,
				kh,
				err,
			)
		}
		return nil, fmt.Errorf("ssh handshake gagal ke %s: %w", sshAddr, err)
	}
	client := ssh.NewClient(cc, chans, reqs)

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		_ = client.Close()
		cancel()
		return nil, fmt.Errorf("gagal listen local port 127.0.0.1:%d: %w", localPort, err)
	}

	t := &SSHTunnel{LocalPort: localPort, listener: ln, sshClient: client, cancel: cancel}

	remoteAddr := net.JoinHostPort(opts.RemoteHost, strconv.Itoa(opts.RemotePort))
	go func() {
		for {
			conn, aerr := ln.Accept()
			if aerr != nil {
				select {
				case <-ctx2.Done():
					return
				default:
					return
				}
			}
			go func(c net.Conn) {
				defer c.Close()
				rc, derr := client.Dial("tcp", remoteAddr)
				if derr != nil {
					return
				}
				defer rc.Close()

				// Bidirectional copy
				done := make(chan struct{}, 2)
				go func() { _, _ = io.Copy(rc, c); done <- struct{}{} }()
				go func() { _, _ = io.Copy(c, rc); done <- struct{}{} }()
				<-done
			}(conn)
		}
	}()

	return t, nil
}

// Stop menghentikan SSH tunnel.
func (t *SSHTunnel) Stop(ctx context.Context) error {
	if t == nil {
		return nil
	}
	t.once.Do(func() {
		if t.cancel != nil {
			t.cancel()
		}
		if t.listener != nil {
			_ = t.listener.Close()
		}
		if t.sshClient != nil {
			_ = t.sshClient.Close()
		}
		select {
		case <-time.After(50 * time.Millisecond):
		case <-ctx.Done():
		}
	})
	return nil
}
