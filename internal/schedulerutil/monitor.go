// File : internal/schedulerutil/monitor.go
// Deskripsi : Helper monitoring systemd units/timers (status/logs/list)
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-04
// Last Modified : 2026-01-04

package schedulerutil

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func isExitCode(err error, code int) bool {
	ee, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}
	return ee.ExitCode() == code
}

type Scope string

const (
	ScopeAuto   Scope = "auto"
	ScopeUser   Scope = "user"
	ScopeSystem Scope = "system"
	ScopeBoth   Scope = "both"
)

func NormalizeScope(v string) (Scope, error) {
	s := Scope(strings.ToLower(strings.TrimSpace(v)))
	switch s {
	case "":
		return ScopeAuto, nil
	case ScopeAuto, ScopeUser, ScopeSystem, ScopeBoth:
		return s, nil
	default:
		return "", fmt.Errorf("scope tidak valid: %s (pilih: auto|user|system|both)", v)
	}
}

func scopesToTry(scope Scope) []Scope {
	// Auto = coba user lalu system. Ini memudahkan user tanpa perlu mikir.
	if scope == ScopeAuto || scope == ScopeBoth {
		return []Scope{ScopeUser, ScopeSystem}
	}
	return []Scope{scope}
}

func systemctlOutput(ctx context.Context, scope Scope, args ...string) (string, error) {
	base := []string{}
	if scope == ScopeUser {
		base = append(base, "--user")
	}
	all := append(base, args...)
	cmd := exec.CommandContext(ctx, "systemctl", all...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func journalctlOutput(ctx context.Context, scope Scope, args ...string) (string, error) {
	base := []string{"--no-pager"}
	if scope == ScopeUser {
		base = append([]string{"--user", "--no-pager"}, args...)
		cmd := exec.CommandContext(ctx, "journalctl", base...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	all := append(base, args...)
	cmd := exec.CommandContext(ctx, "journalctl", all...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// ListServices mengembalikan output systemctl list-units untuk sfdbtools services.
func ListServices(ctx context.Context, scope Scope) (string, error) {
	args := []string{
		"list-units",
		"--all",
		"--type=service",
		"--no-pager",
		"--no-legend",
		"sfdbtools-*.service",
		"sfdbtools-backup@*.service",
	}
	return systemctlOutput(ctx, scope, args...)
}

// ListTimers mengembalikan output systemctl list-timers untuk sfdbtools timers.
func ListTimers(ctx context.Context, scope Scope) (string, error) {
	args := []string{
		"list-timers",
		"--all",
		"--no-pager",
		"--no-legend",
		"sfdbtools-*.timer",
		"sfdbtools-backup@*.timer",
	}
	return systemctlOutput(ctx, scope, args...)
}

// StatusUnit menjalankan systemctl status untuk unit tertentu.
func StatusUnit(ctx context.Context, scope Scope, unit string) (string, error) {
	unit = strings.TrimSpace(unit)
	if unit == "" {
		return "", fmt.Errorf("unit tidak boleh kosong")
	}
	out, err := systemctlOutput(ctx, scope, "status", "--no-pager", unit)
	if err != nil {
		// systemctl status mengembalikan exit code 3 jika unit tidak sedang running.
		// Ini bukan error fatal untuk kebutuhan monitoring (kita tetap ingin menampilkan output status).
		if ee, ok := err.(*exec.ExitError); ok {
			if ee.ExitCode() == 3 {
				return out, nil
			}
		}
	}
	return out, err
}

// StopUnit menjalankan systemctl stop untuk unit tertentu.
func StopUnit(ctx context.Context, scope Scope, unit string) (string, error) {
	unit = strings.TrimSpace(unit)
	if unit == "" {
		return "", fmt.Errorf("unit tidak boleh kosong")
	}
	out, err := systemctlOutput(ctx, scope, "stop", unit)
	// exit code 5 = unit tidak ada / tidak loaded. Untuk operasi stop, ini bisa dianggap idempotent.
	if err != nil && isExitCode(err, 5) {
		return out, nil
	}
	return out, err
}

// DisableUnit menjalankan systemctl disable --now untuk unit tertentu.
func DisableUnit(ctx context.Context, scope Scope, unit string) (string, Scope, error) {
	unit = strings.TrimSpace(unit)
	if unit == "" {
		return "", "", fmt.Errorf("unit tidak boleh kosong")
	}
	out, used, err := TryWithScopes(scope, func(s Scope) (string, error) {
		out, err := systemctlOutput(ctx, s, "disable", "--now", unit)
		if err != nil && isExitCode(err, 5) {
			return out, nil
		}
		return out, err
	})
	return out, used, err
}

// ResetFailedUnit menjalankan systemctl reset-failed untuk unit tertentu.
func ResetFailedUnit(ctx context.Context, scope Scope, unit string) (string, Scope, error) {
	unit = strings.TrimSpace(unit)
	if unit == "" {
		return "", "", fmt.Errorf("unit tidak boleh kosong")
	}
	out, used, err := TryWithScopes(scope, func(s Scope) (string, error) {
		out, err := systemctlOutput(ctx, s, "reset-failed", unit)
		if err != nil && isExitCode(err, 5) {
			return out, nil
		}
		return out, err
	})
	return out, used, err
}

// DaemonReload menjalankan systemctl daemon-reload.
func DaemonReload(ctx context.Context, scope Scope) (string, Scope, error) {
	out, used, err := TryWithScopes(scope, func(s Scope) (string, error) {
		return systemctlOutput(ctx, s, "daemon-reload")
	})
	return out, used, err
}

// LogsUnit menjalankan journalctl untuk unit tertentu.
func LogsUnit(ctx context.Context, scope Scope, unit string, lines int, follow bool) (string, error) {
	unit = strings.TrimSpace(unit)
	if unit == "" {
		return "", fmt.Errorf("unit tidak boleh kosong")
	}
	args := []string{"-u", unit, "--no-pager"}
	if lines > 0 {
		args = append(args, fmt.Sprintf("-n%d", lines))
	}
	if follow {
		args = append(args, "-f")
	}
	return journalctlOutput(ctx, scope, args...)
}

// TryWithScopes mencoba beberapa scope sampai ada yang berhasil.
// Mengembalikan scope yang berhasil dipakai (untuk ditampilkan ke user).
func TryWithScopes[T any](scope Scope, fn func(s Scope) (T, error)) (T, Scope, error) {
	var zero T
	var lastErr error
	for _, sc := range scopesToTry(scope) {
		out, err := fn(sc)
		if err == nil {
			return out, sc, nil
		}
		lastErr = err
	}
	return zero, "", lastErr
}
