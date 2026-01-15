// File : internal/app/profile/process/tunnel.go
// Deskripsi : SSH tunnel native Go (port forwarding)
// Author : Hadiyatna Muflihun
// Tanggal : 2 Januari 2026
// Last Modified : 15 Januari 2026

package process

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

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
	go forward(ctx2, ln, client, remoteAddr)

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
