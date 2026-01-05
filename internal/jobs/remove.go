// File : internal/jobs/remove.go
// Deskripsi : Remove/cleanup unit systemd sfDBTools (stop/disable/reset-failed + optional purge unit file)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026
package jobs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"sfDBTools/internal/services/scheduler"
)

const systemdSystemUnitDir = "/etc/systemd/system"

var allowedUnitRe = regexp.MustCompile(`^(sfdbtools-[a-z0-9_-]+\.(service|timer)|sfdbtools-backup@([a-zA-Z0-9._-]+)?\.(service|timer))$`)

type RemoveOptions struct {
	Scope scheduler.Scope
	Purge bool // delete unit file if exists (system scope only)
}

// Remove menghentikan unit (atau disable timer) dan membersihkan state failed.
// Jika Purge=true dan scope=system (root), unit file di /etc/systemd/system akan dihapus bila ada.
func Remove(ctx context.Context, unit string, opts RemoveOptions) (string, scheduler.Scope, error) {
	unit = strings.TrimSpace(unit)
	if unit == "" {
		return "", "", fmt.Errorf("unit tidak boleh kosong")
	}
	if !allowedUnitRe.MatchString(unit) {
		return "", "", fmt.Errorf("unit tidak diizinkan untuk dihapus: %s", unit)
	}
	if opts.Purge && opts.Scope != scheduler.ScopeSystem {
		return "", "", fmt.Errorf("--purge hanya didukung untuk --scope=system")
	}

	// Stop/disable + reset-failed, dengan behavior idempotent.
	var output strings.Builder
	usedScope := scheduler.Scope("")

	// Disable timer jika timer; stop jika service.
	if strings.HasSuffix(unit, ".timer") {
		out, used, err := scheduler.DisableUnit(ctx, opts.Scope, unit)
		output.WriteString(out)
		if err != nil {
			return output.String(), used, err
		}
		usedScope = used
	} else {
		out, used, err := scheduler.TryWithScopes(opts.Scope, func(s scheduler.Scope) (string, error) {
			return scheduler.StopUnit(ctx, s, unit)
		})
		output.WriteString(out)
		if err != nil {
			return output.String(), used, err
		}
		usedScope = used
	}

	// Reset failed state biar tidak "failed" terus di list.
	out, _, _ := scheduler.ResetFailedUnit(ctx, usedScope, unit)
	if strings.TrimSpace(out) != "" {
		if output.Len() > 0 {
			output.WriteString("\n")
		}
		output.WriteString(out)
	}

	// Optional purge unit file (system scope only)
	if opts.Purge {
		if os.Geteuid() != 0 {
			return output.String(), usedScope, fmt.Errorf("--purge butuh root (gunakan sudo)")
		}
		path := filepath.Join(systemdSystemUnitDir, unit)
		if _, err := os.Stat(path); err == nil {
			if rmErr := os.Remove(path); rmErr != nil {
				return output.String(), usedScope, fmt.Errorf("gagal hapus unit file %s: %w", path, rmErr)
			}
			// Reload systemd setelah hapus file
			_, _, _ = scheduler.DaemonReload(ctx, scheduler.ScopeSystem)
		}
	}

	return output.String(), usedScope, nil
}
