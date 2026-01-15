package connection
// File : internal/app/profile/connection/connector.go
// Deskripsi : Koneksi database berbasis ProfileInfo
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package connection

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	profileerrors "sfdbtools/internal/app/profile/errors"
	"sfdbtools/internal/app/profile/process"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/ui/progress"

	"github.com/mattn/go-isatty"
)

// ConnectWithProfile membuat koneksi database menggunakan ProfileInfo.
func ConnectWithProfile(profile *domain.ProfileInfo, initialDB string) (*database.Client, error) {
	if profile == nil {
		return nil, profileerrors.ErrProfileNil
	}
	if err := ValidateConnectPreflight(profile); err != nil {
		return nil, err
	}

	if initialDB == "" {
		initialDB = consts.DefaultInitialDatabase
	}

	// Spinner message: tampilkan mode koneksi (direct vs SSH tunnel)
	// Non-interaktif (bukan TTY) diperlakukan sama seperti quiet untuk mencegah output spinner merusak pipeline.
	quiet := runtimecfg.IsQuiet() || runtimecfg.IsDaemon() || !isatty.IsTerminal(os.Stdin.Fd()) || !isatty.IsTerminal(os.Stdout.Fd())

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

		ctx, cancel := context.WithTimeout(context.Background(), ProfileConnectTimeout())
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
			ConnectTimeout: ProfileConnectTimeout(),
			ServerAlive:    30 * time.Second,
			ExitOnFailure:  true,
			BatchMode:      true,
		})
		if err != nil {
			sshPort := profile.SSHTunnel.Port
			if sshPort == 0 {
				sshPort = 22
			}
			return nil, fmt.Errorf("gagal membuat SSH tunnel ke %s:%d: %w", sshHost, sshPort, err)
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

	client, err := database.NewClient(context.Background(), cfg, ProfileConnectTimeout(), 10, 5, 0)
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

// TrimProfileSuffix menghapus suffix ekstensi profile (.cnf/.enc) dari nama jika ada.
func TrimProfileSuffix(name string) string {
	n := strings.TrimSpace(name)
	n = strings.TrimSuffix(n, consts.ExtEnc)
	n = strings.TrimSuffix(n, consts.ExtCnf)
	return n
}
