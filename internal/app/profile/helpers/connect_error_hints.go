// File : internal/app/profile/helpers/connect_error_hints.go
// Deskripsi : Helper untuk merangkum error koneksi DB/SSH dengan hint yang actionable
// Author : Hadiyatna Muflihun
// Tanggal : 9 Januari 2026
// Last Modified : 9 Januari 2026

package helpers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

type ConnectErrorKind string

const (
	ConnectErrorKindSSH ConnectErrorKind = "ssh"
	ConnectErrorKindDB  ConnectErrorKind = "db"
)

type ConnectErrorInfo struct {
	Kind   ConnectErrorKind
	Title  string
	Detail string
	Hints  []string
}

func DescribeConnectError(err error) ConnectErrorInfo {
	if err == nil {
		return ConnectErrorInfo{}
	}

	msg := strings.TrimSpace(err.Error())
	lower := strings.ToLower(msg)

	info := ConnectErrorInfo{
		Kind:   ConnectErrorKindDB,
		Title:  "Koneksi database gagal",
		Detail: summarizeMessage(msg, 10, 800),
		Hints:  nil,
	}

	if looksLikeSSHTunnelError(lower) {
		info.Kind = ConnectErrorKindSSH
		info.Title = "SSH tunnel gagal"
	}

	if isTimeoutError(err) {
		t := ProfileConnectTimeout()
		label := "koneksi"
		if info.Kind == ConnectErrorKindSSH {
			label = "SSH"
		} else if info.Kind == ConnectErrorKindDB {
			label = "database"
		}
		info.Hints = append(info.Hints, fmt.Sprintf("Timeout %s (batas: %s). Jika jaringan lambat/VPN putus, cek konektivitas dan coba lagi.", label, formatDurationShort(t)))
	}

	info.Hints = append(info.Hints, classifyHints(lower)...)
	return info
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	// Context deadline
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	// net.Error timeout
	var ne net.Error
	if errors.As(err, &ne) {
		return ne.Timeout()
	}
	// Fallback: pesan error
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "i/o timeout") || strings.Contains(lower, "timeout")
}

func formatDurationShort(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	// Round to seconds for nicer UX
	sec := d.Round(time.Second)
	return sec.String()
}

func looksLikeSSHTunnelError(lower string) bool {
	sshMarkers := []string{
		"ssh handshake",
		"known_hosts",
		"identity file",
		"metode autentikasi ssh",
		"konek tcp ke ssh",
		"ssh tunnel",
		"unable to authenticate",
		"no supported methods remain",
		"permission denied",
	}
	for _, m := range sshMarkers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

func classifyHints(lower string) []string {
	var hints []string

	// =============================================================================
	// SSH tunnel
	// =============================================================================
	if strings.Contains(lower, "metode autentikasi ssh tidak tersedia") {
		hints = append(hints,
			"Isi SSH password, atau set SSH identity file, atau pastikan ssh-agent aktif (SSH_AUTH_SOCK).",
		)
	}
	if strings.Contains(lower, "gagal membaca identity file") || strings.Contains(lower, "no such file") {
		hints = append(hints,
			"Cek path identity file, pastikan file key ada dan readable oleh user yang menjalankan sfdbtools.",
		)
	}
	if strings.Contains(lower, "gagal parse identity file") {
		hints = append(hints,
			"Private key tidak valid/unsupported atau butuh passphrase; coba gunakan key tanpa passphrase atau pakai ssh-agent.",
		)
	}
	if strings.Contains(lower, "unprotected private key file") || strings.Contains(lower, "bad permissions") {
		hints = append(hints,
			"Permission private key terlalu longgar; coba set: chmod 600 <identity_file> (dan pastikan owner benar).",
		)
	}
	if strings.Contains(lower, "host key verification failed") {
		hints = append(hints,
			"Verifikasi host key gagal; pastikan host benar dan update entry di known_hosts yang dipakai sfdbtools.",
		)
	}
	if strings.Contains(lower, "known_hosts key mismatch") {
		hints = append(hints,
			"Host key SSH berubah/mismatch; pastikan host benar lalu update entry di file known_hosts yang disebut.",
		)
	}
	if strings.Contains(lower, "unable to authenticate") || strings.Contains(lower, "no supported methods remain") || strings.Contains(lower, "permission denied") {
		hints = append(hints,
			"Cek SSH user/password/key dan pastikan key terdaftar di ~/.ssh/authorized_keys pada server bastion.",
		)
	}
	if strings.Contains(lower, "could not resolve") || strings.Contains(lower, "no such host") {
		hints = append(hints,
			"Cek DNS/hostname SSH host (bastion) dan konektivitas jaringan.",
		)
	}
	if strings.Contains(lower, "connection refused") {
		hints = append(hints,
			"Cek SSH port (default 22), firewall/security group, dan pastikan service SSH aktif.",
		)
	}
	if strings.Contains(lower, "i/o timeout") || strings.Contains(lower, "timeout") {
		hints = append(hints,
			"Timeout koneksi; cek jaringan/VPN, firewall, dan coba pastikan host/port benar.",
		)
	}
	if strings.Contains(lower, "identity file ssh") {
		hints = append(hints,
			"Jika memakai path relatif untuk identity file, path dihitung relatif dari working directory saat sfdbtools dijalankan.",
		)
	}

	// =============================================================================
	// Database
	// =============================================================================
	if strings.Contains(lower, "dial tcp") {
		hints = append(hints,
			"Cek host/port database dan konektivitas jaringan. Jika memakai SSH tunnel, pastikan tunnel berhasil terbentuk.",
		)
	}
	if strings.Contains(lower, "connection refused") {
		hints = append(hints,
			"Connection refused; cek service database aktif, port benar, dan firewall/security group.",
		)
	}
	if strings.Contains(lower, "no such host") {
		hints = append(hints,
			"Hostname database tidak bisa di-resolve; cek DNS atau gunakan IP.",
		)
	}
	if strings.Contains(lower, "access denied") || strings.Contains(lower, "error 1045") {
		hints = append(hints,
			"Cek DB user/password, allowlist host, dan privilege user pada server database.",
		)
	}
	if strings.Contains(lower, "unknown database") {
		hints = append(hints,
			"Database awal tidak ditemukan atau user tidak punya akses; cek permission user atau target database.",
		)
	}
	if strings.Contains(lower, "x509") || strings.Contains(lower, "tls") || strings.Contains(lower, "ssl") {
		hints = append(hints,
			"Ada isu TLS/SSL; cek konfigurasi SSL di server (sertifikat/CA) dan kebijakan koneksi.",
		)
	}
	if strings.Contains(lower, "auth_gssapi_client") {
		hints = append(hints,
			"Server memakai auth plugin yang tidak didukung client ini; pertimbangkan ubah auth user ke mysql_native_password/caching_sha2_password.",
		)
	}
	if strings.Contains(lower, "caching_sha2_password") {
		hints = append(hints,
			"Auth plugin caching_sha2_password terdeteksi; pastikan koneksi aman (TLS) atau gunakan auth plugin yang kompatibel.",
		)
	}
	if strings.Contains(lower, "public key retrieval is not allowed") {
		hints = append(hints,
			"Server meminta public key retrieval; biasanya perlu TLS atau opsi allowPublicKeyRetrieval (MySQL 8). Pertimbangkan gunakan TLS atau ubah auth plugin user.",
		)
	}
	if strings.Contains(lower, "authentication protocol") || strings.Contains(lower, "client does not support") {
		hints = append(hints,
			"Ada mismatch protokol/auth; coba ubah plugin auth user (mis. mysql_native_password) atau pastikan client/driver kompatibel dengan versi server.",
		)
	}
	if strings.Contains(lower, "host database kosong") || strings.Contains(lower, "user database kosong") || strings.Contains(lower, "port database tidak valid") {
		hints = append(hints,
			"Preflight gagal; perbaiki field yang kosong/tidak valid lalu coba lagi.",
		)
	}

	return uniqueHints(hints)
}

func uniqueHints(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, h := range in {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		if _, ok := seen[h]; ok {
			continue
		}
		seen[h] = struct{}{}
		out = append(out, h)
	}
	return out
}

func summarizeMessage(msg string, maxLines int, maxChars int) string {
	if strings.TrimSpace(msg) == "" {
		return ""
	}
	lines := strings.Split(msg, "\n")
	if maxLines > 0 && len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	s := strings.TrimSpace(strings.Join(lines, "\n"))
	if maxChars > 0 && len(s) > maxChars {
		s = s[:maxChars] + "..."
	}
	return s
}
