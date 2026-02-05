package connection

// File : internal/app/profile/connection/connector.go
// Deskripsi : Koneksi database berbasis ProfileInfo
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

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

// ConnectWithProfile membuat koneksi database menggunakan ProfileInfo dengan support untuk direct dan SSH tunnel.
//
// Fungsi ini menangani:
//  1. Pre-connection validation (ValidateConnectPreflight)
//  2. SSH tunnel setup (jika profile.SSHTunnel.Enabled=true)
//  3. Database connection dengan timeout
//  4. Connection lifecycle management (close tunnel saat client close)
//
// Connection Modes:
//   - Direct: Connect langsung ke database host:port
//   - SSH Tunnel: Start SSH tunnel dulu, lalu connect ke localhost:localPort
//
// Parameters:
//   - cfg: Config object yang menyediakan ProfileConnectTimeout (via interface)
//   - profile: ProfileInfo yang berisi DB credentials dan SSH tunnel config
//   - initialDB: Initial database untuk connection (default: "information_schema")
//
// SSH Tunnel Behavior:
//   - Tunnel di-start dengan SSH command (via process.StartSSHTunnel)
//   - Local port: profile.SSHTunnel.LocalPort (0 = auto assign)
//   - Tunnel parameters: ConnectTimeout, ServerAlive, BatchMode, ExitOnFailure
//   - Authentication: Password atau IdentityFile (private key)
//   - Resolved local port disimpan di profile.SSHTunnel.ResolvedLocalPort
//
// Returns:
//   - *database.Client: Connected database client dengan lifecycle hooks
//   - error: Connection error (DNS, TCP, SSH tunnel, auth, dll)
//
// Error Handling:
//   - Pre-flight validation errors (missing required fields)
//   - SSH tunnel errors (connection failed, auth failed, remote host unreachable)
//   - Database connection errors (timeout, auth failed, host unreachable)
//   - Tunnel cleanup pada connection error (no resource leak)
//
// Client Lifecycle:
//   - Client.SetOnClose() registered untuk cleanup SSH tunnel
//   - Saat client.Close() dipanggil, tunnel akan di-stop otomatis
//   - Proper cleanup meskipun program terminated (context cancellation)
//
// Display Progress:
//   - Show spinner dengan elapsed time (untuk interactive mode)
//   - Quiet mode: no spinner (untuk automation/CI)
//   - Non-TTY: no spinner (untuk pipeline)
//
// Example:
//
//	client, err := connection.ConnectWithProfile(cfg, profile, "information_schema")
//	if err != nil {
//		log.Fatalf("Connection failed: %v", err)
//	}
//	defer client.Close() // Auto cleanup tunnel jika ada
//
// Security Notes:
//   - Password tidak di-log
//   - SSH key path validated (file exists check)
//   - Known hosts verification (optional, default=false untuk automation)
func ConnectWithProfile(cfg interface{}, profile *domain.ProfileInfo, initialDB string) (*database.Client, error) {
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
	quiet := runtimecfg.IsQuiet() || !isatty.IsTerminal(os.Stdin.Fd()) || !isatty.IsTerminal(os.Stdout.Fd())

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

		ctx, cancel := context.WithTimeout(context.Background(), ProfileConnectTimeout(cfg))
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
			ConnectTimeout: ProfileConnectTimeout(cfg),
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
	dbCfg := database.Config{
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

	client, err := database.NewClient(context.Background(), dbCfg, ProfileConnectTimeout(cfg), 10, 5, 0)
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
//
// Fungsi ini membersihkan nama profile dari ekstensi file untuk display dan comparison:
//   - "prod-db.cnf.enc" → "prod-db"
//   - "prod-db.cnf" → "prod-db"
//   - "prod-db.enc" → "prod-db"
//   - "prod-db" → "prod-db" (no change)
//
// Trim dilakukan berurutan: .enc dulu, lalu .cnf (handle double extension).
//
// Use case:
//   - Display profile name tanpa ekstensi (user-facing)
//   - Profile name comparison (ignore extension)
//   - Build file name (re-add extension setelah sanitize)
func TrimProfileSuffix(name string) string {
	n := strings.TrimSpace(name)
	n = strings.TrimSuffix(n, consts.ExtEnc)
	n = strings.TrimSuffix(n, consts.ExtCnf)
	return n
}
