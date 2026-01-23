// File : internal/app/profile/connection/test.go
// Deskripsi : Connection test detail untuk profile (DNS/TCP/SSH/DB)
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026

package connection

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"sfdbtools/internal/app/profile/process"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
)

type StepStatus string

const (
	StepStatusSuccess  StepStatus = "success"
	StepStatusFailed   StepStatus = "failed"
	StepStatusSkipped  StepStatus = "skipped"
	StepStatusDisabled StepStatus = "disabled"
)

type StepResult struct {
	Status   StepStatus
	Duration time.Duration
	Detail   string
	Err      error
}

func (s StepResult) Display() string {
	switch s.Status {
	case StepStatusSuccess:
		if s.Duration > 0 {
			return fmt.Sprintf("Success (%s)", formatDurationMs(s.Duration))
		}
		return "Success"
	case StepStatusSkipped:
		if strings.TrimSpace(s.Detail) != "" {
			return "Skipped (" + s.Detail + ")"
		}
		return "Skipped"
	case StepStatusDisabled:
		return "Disabled"
	case StepStatusFailed:
		if strings.TrimSpace(s.Detail) != "" {
			return "Failed (" + s.Detail + ")"
		}
		return "Failed"
	default:
		return "-"
	}
}

type ConnectionTestReport struct {
	DNSResolution  StepResult
	TCPConnection  StepResult
	SSHTunnel      StepResult
	Authentication StepResult
	DBVersion      string
	TotalLatency   time.Duration
	Healthy        bool

	// Err berisi error utama yang paling relevan untuk hint.
	Err error
}

// TestConnection menjalankan serangkaian test koneksi yang lebih detail.
// Catatan:
// - Untuk SSH tunnel: DNS/TCP test diarahkan ke SSH host.
// - Untuk direct: DNS/TCP test diarahkan ke DB host.
// - Fungsi ini tidak menampilkan spinner; cocok untuk dipanggil dari UI summary/table.
func TestConnection(cfg interface{}, profile *domain.ProfileInfo, initialDB string) *ConnectionTestReport {
	report := &ConnectionTestReport{}
	start := time.Now()
	defer func() {
		report.TotalLatency = time.Since(start)
		report.Healthy = report.Err == nil
	}()

	if profile == nil {
		report.Err = fmt.Errorf("profile nil")
		report.DNSResolution = StepResult{Status: StepStatusFailed, Detail: "profile nil", Err: report.Err}
		report.TCPConnection = StepResult{Status: StepStatusDisabled}
		report.SSHTunnel = StepResult{Status: StepStatusDisabled}
		report.Authentication = StepResult{Status: StepStatusDisabled}
		return report
	}

	// Preflight tetap dijalankan agar error invalid input tampil lebih cepat.
	if err := ValidateConnectPreflight(profile); err != nil {
		report.Err = err
		report.DNSResolution = StepResult{Status: StepStatusFailed, Detail: "preflight", Err: err}
		report.TCPConnection = StepResult{Status: StepStatusDisabled}
		report.SSHTunnel = StepResult{Status: StepStatusDisabled}
		report.Authentication = StepResult{Status: StepStatusDisabled}
		return report
	}

	timeout := ProfileConnectTimeout(cfg)
	if initialDB == "" {
		initialDB = consts.DefaultInitialDatabase
	}

	// Tentukan target DNS/TCP: SSH host (jika tunnel), else DB host.
	dnsHost := strings.TrimSpace(profile.DBInfo.Host)
	tcpHost := dnsHost
	tcpPort := profile.DBInfo.Port
	if profile.SSHTunnel.Enabled {
		sshHost := strings.TrimSpace(profile.SSHTunnel.Host)
		if sshHost != "" {
			dnsHost = sshHost
			tcpHost = sshHost
		}
		sshPort := profile.SSHTunnel.Port
		if sshPort == 0 {
			sshPort = 22
		}
		tcpPort = sshPort
	}

	// Step 1: DNS resolution
	report.DNSResolution = testDNS(timeout, dnsHost)
	if report.DNSResolution.Status == StepStatusFailed {
		report.Err = report.DNSResolution.Err
		report.TCPConnection = StepResult{Status: StepStatusDisabled}
		report.SSHTunnel = StepResult{Status: StepStatusDisabled}
		report.Authentication = StepResult{Status: StepStatusDisabled}
		return report
	}

	// Step 2: TCP connection
	report.TCPConnection = testTCP(timeout, tcpHost, tcpPort)
	if report.TCPConnection.Status == StepStatusFailed {
		report.Err = report.TCPConnection.Err
		report.SSHTunnel = StepResult{Status: StepStatusDisabled}
		report.Authentication = StepResult{Status: StepStatusDisabled}
		return report
	}

	// Step 3: SSH tunnel (optional)
	var tunnel *process.SSHTunnel
	if profile.SSHTunnel.Enabled {
		tunnelStart := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		t, err := process.StartSSHTunnel(ctx, process.SSHTunnelOptions{
			SSHHost:        strings.TrimSpace(profile.SSHTunnel.Host),
			SSHPort:        profile.SSHTunnel.Port,
			SSHUser:        profile.SSHTunnel.User,
			Password:       profile.SSHTunnel.Password,
			IdentityFile:   profile.SSHTunnel.IdentityFile,
			LocalPort:      profile.SSHTunnel.LocalPort,
			RemoteHost:     profile.DBInfo.Host,
			RemotePort:     profile.DBInfo.Port,
			ConnectTimeout: timeout,
			ServerAlive:    30 * time.Second,
			ExitOnFailure:  true,
			BatchMode:      true,
		})
		if err != nil {
			report.SSHTunnel = StepResult{Status: StepStatusFailed, Duration: time.Since(tunnelStart), Detail: "ssh tunnel", Err: err}
			report.Err = fmt.Errorf("ssh tunnel gagal: %w", err)
			report.Authentication = StepResult{Status: StepStatusDisabled}
			return report
		}
		tunnel = t
		profile.SSHTunnel.ResolvedLocalPort = tunnel.LocalPort
		report.SSHTunnel = StepResult{Status: StepStatusSuccess, Duration: time.Since(tunnelStart), Detail: fmt.Sprintf("Active (Local Port: %d)", tunnel.LocalPort)}
	} else {
		report.SSHTunnel = StepResult{Status: StepStatusDisabled}
	}

	// Step 4: DB authentication (connect + ping)
	authStart := time.Now()
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

	client, err := database.NewClient(context.Background(), dbCfg, timeout, 2, 1, 0)
	if err != nil {
		if tunnel != nil {
			_ = tunnel.Stop(context.Background())
		}
		report.Authentication = StepResult{Status: StepStatusFailed, Duration: time.Since(authStart), Detail: "db auth", Err: err}
		report.Err = err
		return report
	}
	defer func() {
		_ = client.Close()
		if tunnel != nil {
			_ = tunnel.Stop(context.Background())
		}
	}()
	report.Authentication = StepResult{Status: StepStatusSuccess, Duration: time.Since(authStart)}

	// Step 5: DB version
	verCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ver, verr := client.GetVersion(verCtx)
	if verr != nil {
		report.DBVersion = "-"
		// Tidak fatal untuk health; tapi tetap catat error agar user aware.
		if report.Err == nil {
			report.Err = verr
		}
		return report
	}
	report.DBVersion = ver

	return report
}

func testDNS(timeout time.Duration, host string) StepResult {
	h := strings.TrimSpace(host)
	if h == "" {
		err := fmt.Errorf("host kosong")
		return StepResult{Status: StepStatusFailed, Detail: "host kosong", Err: err}
	}

	// Jika IP address, skip lookup (lebih cepat dan menghindari false negative DNS).
	if net.ParseIP(h) != nil {
		return StepResult{Status: StepStatusSuccess, Duration: 0, Detail: "ip"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	_, err := net.DefaultResolver.LookupHost(ctx, h)
	if err != nil {
		return StepResult{Status: StepStatusFailed, Duration: time.Since(start), Detail: "dns", Err: err}
	}
	return StepResult{Status: StepStatusSuccess, Duration: time.Since(start)}
}

func testTCP(timeout time.Duration, host string, port int) StepResult {
	h := strings.TrimSpace(host)
	if h == "" {
		err := fmt.Errorf("host kosong")
		return StepResult{Status: StepStatusFailed, Detail: "host kosong", Err: err}
	}
	if port <= 0 || port > 65535 {
		err := fmt.Errorf("port tidak valid: %d", port)
		return StepResult{Status: StepStatusFailed, Detail: "port invalid", Err: err}
	}

	addr := net.JoinHostPort(h, strconv.Itoa(port))
	start := time.Now()
	d := net.Dialer{Timeout: timeout}
	conn, err := d.Dial("tcp", addr)
	if err != nil {
		return StepResult{Status: StepStatusFailed, Duration: time.Since(start), Detail: "tcp", Err: err}
	}
	_ = conn.Close()
	return StepResult{Status: StepStatusSuccess, Duration: time.Since(start)}
}

func formatDurationMs(d time.Duration) string {
	if d <= 0 {
		return "0ms"
	}
	// Round to ms for stable display.
	ms := d.Round(time.Millisecond)
	return ms.String()
}
