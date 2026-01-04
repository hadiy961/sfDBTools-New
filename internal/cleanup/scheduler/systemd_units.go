// File : internal/cleanup/scheduler/systemd_units.go
// Deskripsi : Generator unit systemd untuk scheduler cleanup
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-02
// Last Modified : 2026-01-04

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

	content := strings.Join([]string{
		"[Unit]",
		"Description=sfDBTools Cleanup",
		"After=network-online.target",
		"Wants=network-online.target",
		"",
		"[Service]",
		"Type=oneshot",
		fmt.Sprintf("EnvironmentFile=-%s", defaultEnvFile),
		"",
		fmt.Sprintf("ExecStart=/usr/bin/flock -x %s /usr/bin/sfdbtools --daemon %s", defaultLockFile, args),
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
