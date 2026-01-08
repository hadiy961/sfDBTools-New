package helpers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/shared/process"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/ui/progress"
)

// EffectiveDBInfo mengembalikan DBInfo yang efektif untuk koneksi.
// Jika SSH tunnel aktif dan sudah memiliki ResolvedLocalPort, koneksi diarahkan ke localhost.
func EffectiveDBInfo(profile *domain.ProfileInfo) domain.DBInfo {
	if profile == nil {
		return domain.DBInfo{}
	}
	info := profile.DBInfo
	if profile.SSHTunnel.Enabled && profile.SSHTunnel.ResolvedLocalPort > 0 {
		info.Host = "127.0.0.1"
		info.Port = profile.SSHTunnel.ResolvedLocalPort
	}
	return info
}

// ConnectWithProfile membuat koneksi database menggunakan ProfileInfo.
func ConnectWithProfile(profile *domain.ProfileInfo, initialDB string) (*database.Client, error) {
	if profile == nil {
		return nil, fmt.Errorf("profile tidak boleh nil")
	}

	if initialDB == "" {
		initialDB = consts.DefaultInitialDatabase
	}

	// Spinner message: tampilkan mode koneksi (direct vs SSH tunnel)
	quiet := runtimecfg.IsQuiet() || runtimecfg.IsDaemon()

	name := strings.TrimSpace(profile.Name)
	if name == "" {
		name = strings.TrimSpace(profile.DBInfo.Host)
		if name == "" {
			name = "database"
		}
	}

	modeText := "melalui koneksi langsung"
	if profile.SSHTunnel.Enabled {
		modeText = "melalui SSH Tunnel"
	}

	var spin *progress.Spinner
	if !quiet {
		spin = progress.NewSpinnerWithElapsed(fmt.Sprintf("Menghubungkan ke %s %s", name, modeText))
		spin.Start()
		defer spin.Stop()
	}

	// SSH tunnel mode: start tunnel dan arahkan koneksi ke localhost.
	var tunnel *process.SSHTunnel
	if profile.SSHTunnel.Enabled {
		sshHost := strings.TrimSpace(profile.SSHTunnel.Host)
		if sshHost == "" {
			return nil, fmt.Errorf("ssh tunnel aktif tetapi ssh-host kosong")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		t, err := process.StartSSHTunnel(ctx, process.SSHTunnelOptions{
			SSHHost:        sshHost,
			SSHPort:        profile.SSHTunnel.Port,
			SSHUser:        profile.SSHTunnel.User,
			Password:       profile.SSHTunnel.Password,
			IdentityFile:   profile.SSHTunnel.IdentityFile,
			LocalPort:      profile.SSHTunnel.LocalPort,
			RemoteHost:     profile.DBInfo.Host,
			RemotePort:     profile.DBInfo.Port,
			ConnectTimeout: 10 * time.Second,
			ServerAlive:    30 * time.Second,
			ExitOnFailure:  true,
			BatchMode:      true,
		})
		if err != nil {
			return nil, err
		}
		tunnel = t
		profile.SSHTunnel.ResolvedLocalPort = tunnel.LocalPort
	}

	info := EffectiveDBInfo(profile)
	cfg := database.Config{
		Host:                 info.Host,
		Port:                 info.Port,
		User:                 info.User,
		Password:             info.Password,
		AllowNativePasswords: true,
		ParseTime:            true,
		Database:             initialDB,
		ReadTimeout:          0,
		WriteTimeout:         0,
	}

	client, err := database.NewClient(context.Background(), cfg, 10*time.Second, 10, 5, 0)
	if err != nil {
		if tunnel != nil {
			_ = tunnel.Stop(context.Background())
		}
		return nil, fmt.Errorf("gagal koneksi ke %s@%s:%d: %w",
			profile.DBInfo.User, profile.DBInfo.Host, profile.DBInfo.Port, err)
	}

	if tunnel != nil {
		client.SetOnClose(func() error {
			return tunnel.Stop(context.Background())
		})
	}

	return client, nil
}
