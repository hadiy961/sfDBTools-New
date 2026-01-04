// File : internal/cleanup/scheduler/systemd_units.go
// Deskripsi : Generator unit systemd untuk scheduler cleanup
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-02
// Last Modified : 2026-01-05

package scheduler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	appdeps "sfDBTools/internal/deps"
	"sfDBTools/internal/schedulerutil"
)

const (
	systemdUnitDir  = "/etc/systemd/system"
	defaultLockFile = "/var/lock/sfdbtools-backup.lock" // shared lock (serialize with scheduled backups)
	defaultEnvFile  = "/etc/sfDBTools/.env"

	cleanupServiceUnit = "sfdbtools-cleanup.service"
	cleanupTimerUnit   = "sfdbtools-cleanup.timer"
)

func Start(ctx context.Context, deps *appdeps.Dependencies, dryRun bool) error {
	if err := schedulerutil.EnsureLinux(); err != nil {
		return err
	}
	if err := ensureRoot(); err != nil {
		return err
	}
	if err := ensureEnvFileSecure(defaultEnvFile); err != nil {
		return err
	}
	enabled, schedule, err := getCleanupConfig(deps)
	if err != nil {
		return err
	}
	if !enabled {
		return fmt.Errorf("cleanup scheduler tidak aktif: set backup.cleanup.enabled=true")
	}
	if strings.TrimSpace(schedule) == "" {
		return fmt.Errorf("cleanup scheduler butuh schedule: set backup.cleanup.schedule (cron 5 kolom)")
	}

	if err := writeServiceUnit(dryRun); err != nil {
		return err
	}
	if err := writeTimerUnit(schedule); err != nil {
		return err
	}

	if err := schedulerutil.Systemctl(ctx, "daemon-reload"); err != nil {
		return err
	}
	if err := schedulerutil.Systemctl(ctx, "enable", "--now", cleanupTimerUnit); err != nil {
		return err
	}
	return nil
}

func Stop(ctx context.Context, _ *appdeps.Dependencies, killRunning bool) error {
	if err := schedulerutil.EnsureLinux(); err != nil {
		return err
	}
	if err := ensureRoot(); err != nil {
		return err
	}
	_ = schedulerutil.Systemctl(ctx, "disable", "--now", cleanupTimerUnit)
	if killRunning {
		_ = schedulerutil.Systemctl(ctx, "stop", cleanupServiceUnit)
	}
	return nil
}

func Status(ctx context.Context, deps *appdeps.Dependencies) error {
	if err := schedulerutil.EnsureLinux(); err != nil {
		return err
	}
	enabledCfg, schedule, err := getCleanupConfig(deps)
	if err != nil {
		return err
	}

	active, _ := schedulerutil.SystemctlOutput(ctx, "is-active", cleanupTimerUnit)
	enabled, _ := schedulerutil.SystemctlOutput(ctx, "is-enabled", cleanupTimerUnit)
	deps.Logger.Info("Status scheduler cleanup (systemd timer)")
	deps.Logger.Infof("- config.enabled=%v | schedule=%s | systemd: %s/%s", enabledCfg, strings.TrimSpace(schedule), strings.TrimSpace(enabled), strings.TrimSpace(active))
	return nil
}

func getCleanupConfig(deps *appdeps.Dependencies) (bool, string, error) {
	if deps == nil || deps.Config == nil {
		return false, "", fmt.Errorf("config belum tersedia")
	}
	return deps.Config.Backup.Cleanup.Enabled, deps.Config.Backup.Cleanup.Schedule, nil
}

func writeServiceUnit(dryRun bool) error {
	args := "cleanup run"
	if dryRun {
		args += " --dry-run"
	}

	binaryPath := detectBinaryPathForSystemd()

	content := strings.Join([]string{
		"[Unit]",
		"Description=sfDBTools Cleanup",
		"After=network-online.target",
		"Wants=network-online.target",
		"",
		"[Service]",
		"Type=oneshot",
		"UMask=0077",
		"SyslogIdentifier=sfdbtools",
		fmt.Sprintf("EnvironmentFile=-%s", defaultEnvFile),
		"",
		fmt.Sprintf("ExecStart=/usr/bin/flock -x %s %s --daemon %s", defaultLockFile, binaryPath, args),
		"TimeoutStartSec=0",
		"",
		"[Install]",
		"WantedBy=multi-user.target",
	}, "\n") + "\n"

	path := filepath.Join(systemdUnitDir, cleanupServiceUnit)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("gagal menulis unit service %s: %w", path, err)
	}
	return nil
}

func writeTimerUnit(cron string) error {
	onCalendar, err := schedulerutil.CronToOnCalendar(cron)
	if err != nil {
		return fmt.Errorf("schedule cleanup tidak valid: %w", err)
	}

	content := strings.Join([]string{
		"[Unit]",
		"Description=sfDBTools Cleanup Timer",
		"",
		"[Timer]",
		fmt.Sprintf("OnCalendar=%s", onCalendar),
		"Persistent=true",
		fmt.Sprintf("Unit=%s", cleanupServiceUnit),
		"",
		"[Install]",
		"WantedBy=timers.target",
	}, "\n") + "\n"

	path := filepath.Join(systemdUnitDir, cleanupTimerUnit)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("gagal menulis unit timer %s: %w", path, err)
	}
	return nil
}

func ensureRoot() error {
	// Butuh akses tulis /etc/systemd/system dan systemctl enable/disable.
	if os.Geteuid() != 0 {
		return fmt.Errorf("perintah ini butuh root (gunakan sudo)")
	}
	return nil
}

func detectBinaryPathForSystemd() string {
	// Default: sesuai scripts/build_run.sh (dan tar installer).
	defaultPath := "/usr/bin/sfDBTools"
	if _, err := os.Stat("/usr/bin/sfdbtools"); err == nil {
		return "/usr/bin/sfdbtools"
	}
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}
	// Jika belum terinstall, tetap tulis defaultPath agar konsisten.
	return defaultPath
}

func ensureEnvFileSecure(path string) error {
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("gagal cek env file %s: %w", path, err)
	}
	if st.IsDir() {
		return fmt.Errorf("env file %s tidak valid (directory)", path)
	}
	// Env file bisa berisi key sensitif, jadi permission harus ketat.
	if st.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("permission env file %s tidak aman (%o). Disarankan: chown root:root %s && chmod 600 %s", path, st.Mode().Perm(), path, path)
	}
	return nil
}
