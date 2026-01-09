// File : internal/app/backup/scheduler/systemd_units.go
// Deskripsi : Generator unit systemd untuk scheduler backup
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-02
// Last Modified : 2026-01-06
package scheduler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"

	appdeps "sfdbtools/internal/cli/deps"
	appconfig "sfdbtools/internal/services/config"
	"sfdbtools/internal/services/scheduler"
	"sfdbtools/internal/shared/consts"
)

const (
	systemdUnitDir  = "/etc/systemd/system"
	defaultLockFile = "/var/lock/sfdbtools-backup.lock"
	defaultEnvFile  = "/etc/sfDBTools/.env"
	serviceTemplate = "sfdbtools-backup@.service"
)

func Start(ctx context.Context, deps *appdeps.Dependencies, jobName string) error {
	if err := scheduler.EnsureLinux(); err != nil {
		return err
	}
	if err := ensureRoot(); err != nil {
		return err
	}
	jobs, err := getTargetJobs(deps, jobName, true)
	if err != nil {
		return err
	}
	if len(jobs) == 0 {
		return fmt.Errorf("tidak ada job scheduler yang aktif")
	}
	if err := validateJobsForStart(deps, jobs); err != nil {
		return err
	}

	if err := writeServiceTemplate(); err != nil {
		return err
	}
	for _, job := range jobs {
		if err := writeTimerForJob(job); err != nil {
			return err
		}
	}

	if err := scheduler.Systemctl(ctx, "daemon-reload"); err != nil {
		return err
	}

	for _, job := range jobs {
		timerUnit := fmt.Sprintf("sfdbtools-backup@%s.timer", job.Name)
		// Preventive: start bisa dipanggil berulang kali.
		// - `enable` idempotent
		// - `restart` memastikan perubahan timer (OnCalendar) ter-apply walaupun timer sudah aktif
		if err := scheduler.Systemctl(ctx, "enable", timerUnit); err != nil {
			return err
		}
		if err := scheduler.Systemctl(ctx, "restart", timerUnit); err != nil {
			return err
		}
	}

	return nil
}

func Stop(ctx context.Context, deps *appdeps.Dependencies, jobName string, killRunning bool) error {
	if err := scheduler.EnsureLinux(); err != nil {
		return err
	}
	if err := ensureRoot(); err != nil {
		return err
	}
	if killRunning && strings.TrimSpace(jobName) == "" {
		return fmt.Errorf("--kill-running harus dikombinasi dengan --job")
	}
	jobs, err := getTargetJobs(deps, jobName, true)
	if err != nil {
		return err
	}
	if len(jobs) == 0 {
		return fmt.Errorf("tidak ada job scheduler yang aktif")
	}

	for _, job := range jobs {
		timerUnit := fmt.Sprintf("sfdbtools-backup@%s.timer", job.Name)
		_ = scheduler.Systemctl(ctx, "disable", "--now", timerUnit)
		if killRunning {
			serviceUnit := fmt.Sprintf("sfdbtools-backup@%s.service", job.Name)
			_ = scheduler.Systemctl(ctx, "stop", serviceUnit)
		}
	}

	return nil
}

func Status(ctx context.Context, deps *appdeps.Dependencies, jobName string) error {
	if err := scheduler.EnsureLinux(); err != nil {
		return err
	}
	jobs, err := getTargetJobs(deps, jobName, false)
	if err != nil {
		return err
	}
	if len(jobs) == 0 {
		return fmt.Errorf("tidak ada job scheduler di config")
	}

	deps.Logger.Info("Status scheduler backup (systemd timer)")
	for _, job := range jobs {
		timerUnit := fmt.Sprintf("sfdbtools-backup@%s.timer", job.Name)
		active, _ := scheduler.SystemctlOutput(ctx, "is-active", timerUnit)
		enabled, _ := scheduler.SystemctlOutput(ctx, "is-enabled", timerUnit)
		deps.Logger.Infof("- %s | config.enabled=%v | systemd: %s/%s | schedule=%s", job.Name, job.Enabled, strings.TrimSpace(enabled), strings.TrimSpace(active), job.Schedule)
	}
	return nil
}

func getTargetJobs(deps *appdeps.Dependencies, jobName string, onlyEnabled bool) ([]appconfig.BackupSchedulerJob, error) {
	if deps == nil || deps.Config == nil {
		return nil, fmt.Errorf("config belum tersedia")
	}
	jobs := deps.Config.Backup.Scheduler.Jobs
	var out []appconfig.BackupSchedulerJob

	if strings.TrimSpace(jobName) != "" {
		for _, j := range jobs {
			if j.Name == jobName {
				if onlyEnabled && !j.Enabled {
					return nil, fmt.Errorf("job '%s' tidak enabled di config", jobName)
				}
				return []appconfig.BackupSchedulerJob{j}, nil
			}
		}
		return nil, fmt.Errorf("job '%s' tidak ditemukan di config", jobName)
	}

	for _, j := range jobs {
		if onlyEnabled {
			if j.Enabled {
				out = append(out, j)
			}
		} else {
			out = append(out, j)
		}
	}
	return out, nil
}

func writeServiceTemplate() error {
	binaryPath := detectBinaryPathForSystemd()
	content := strings.Join([]string{
		"[Unit]",
		"Description=sfdbtools Backup Job (%i)",
		"After=network-online.target",
		"Wants=network-online.target",
		"",
		"[Service]",
		"Type=oneshot",
		"UMask=0077",
		"SyslogIdentifier=sfdbtools",
		fmt.Sprintf("EnvironmentFile=-%s", defaultEnvFile),
		"",
		// Global lock: semua job antri dan tidak boleh paralel.
		fmt.Sprintf("ExecStart=/usr/bin/flock -x %s %s --daemon db-backup schedule run --job %%i", defaultLockFile, binaryPath),
		"TimeoutStartSec=0",
		"",
		"[Install]",
		"WantedBy=multi-user.target",
	}, "\n") + "\n"

	path := filepath.Join(systemdUnitDir, serviceTemplate)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("gagal menulis unit service %s: %w", path, err)
	}
	return nil
}

func detectBinaryPathForSystemd() string {
	// Default: sesuai scripts/build_run.sh (dan tar installer).
	defaultPath := "/usr/bin/sfdbtools"
	if _, err := os.Stat("/usr/bin/sfdbtools"); err == nil {
		return "/usr/bin/sfdbtools"
	}
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}
	// Jika belum terinstall, tetap tulis defaultPath agar konsisten.
	return defaultPath
}

func writeTimerForJob(job appconfig.BackupSchedulerJob) error {
	if strings.TrimSpace(job.Name) == "" {
		return fmt.Errorf("job name tidak boleh kosong")
	}
	if strings.TrimSpace(job.Schedule) == "" {
		return fmt.Errorf("schedule untuk job '%s' tidak boleh kosong", job.Name)
	}
	onCalendar, err := scheduler.CronToOnCalendar(job.Schedule)
	if err != nil {
		return fmt.Errorf("schedule job '%s' tidak valid: %w", job.Name, err)
	}

	timerUnitName := fmt.Sprintf("sfdbtools-backup@%s.timer", job.Name)
	serviceUnitName := fmt.Sprintf("sfdbtools-backup@%s.service", job.Name)

	content := strings.Join([]string{
		"[Unit]",
		fmt.Sprintf("Description=sfdbtools Backup Timer (%s)", job.Name),
		"",
		"[Timer]",
		fmt.Sprintf("OnCalendar=%s", onCalendar),
		"Persistent=true",
		fmt.Sprintf("Unit=%s", serviceUnitName),
		"",
		"[Install]",
		"WantedBy=timers.target",
	}, "\n") + "\n"

	path := filepath.Join(systemdUnitDir, timerUnitName)
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

func validateJobsForStart(deps *appdeps.Dependencies, jobs []appconfig.BackupSchedulerJob) error {
	if deps == nil || deps.Config == nil {
		return fmt.Errorf("config belum tersedia")
	}
	if err := ensureEnvFileSecure(defaultEnvFile); err != nil {
		return err
	}

	// Pastikan profile-key tersedia via env OS atau file env systemd.
	sourceProfileKey := lookupEnvOrFile(consts.ENV_SOURCE_PROFILE_KEY, defaultEnvFile)
	if strings.TrimSpace(sourceProfileKey) == "" {
		return fmt.Errorf("profile-key wajib untuk scheduler: set env %s (disarankan di %s)", consts.ENV_SOURCE_PROFILE_KEY, defaultEnvFile)
	}

	// Jika enkripsi backup diaktifkan dan key tidak diset di config, pastikan ada env.
	if deps.Config.Backup.Encryption.Enabled && strings.TrimSpace(deps.Config.Backup.Encryption.Key) == "" {
		backupEncKey := lookupEnvOrFile(consts.ENV_BACKUP_ENCRYPTION_KEY, defaultEnvFile)
		if strings.TrimSpace(backupEncKey) == "" {
			return fmt.Errorf("backup encryption aktif tapi key kosong: set backup.encryption.key atau env %s (disarankan di %s)", consts.ENV_BACKUP_ENCRYPTION_KEY, defaultEnvFile)
		}
	}

	for _, job := range jobs {
		if strings.TrimSpace(job.Name) == "" {
			return fmt.Errorf("job name tidak boleh kosong")
		}
		if !job.Enabled {
			continue
		}
		if strings.TrimSpace(job.Schedule) == "" {
			return fmt.Errorf("schedule untuk job '%s' tidak boleh kosong", job.Name)
		}
		if _, err := scheduler.CronToOnCalendar(job.Schedule); err != nil {
			return fmt.Errorf("schedule job '%s' tidak valid: %w", job.Name, err)
		}

		mode := strings.TrimSpace(job.Mode)
		if mode == "" {
			mode = "separated"
		}
		if mode != "separated" && mode != "combined" {
			return fmt.Errorf("mode job '%s' tidak valid: '%s' (gunakan separated/combined)", job.Name, job.Mode)
		}

		if strings.TrimSpace(job.IncludeFile) == "" {
			return fmt.Errorf("include_file wajib untuk job '%s'", job.Name)
		}
		if err := mustExistFile(job.IncludeFile); err != nil {
			return fmt.Errorf("include_file job '%s' tidak valid: %w", job.Name, err)
		}

		if strings.TrimSpace(job.Profile) == "" {
			return fmt.Errorf("profile wajib untuk job '%s'", job.Name)
		}
		if err := mustExistFile(job.Profile); err != nil {
			return fmt.Errorf("profile job '%s' tidak valid: %w", job.Name, err)
		}
	}

	return nil
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

func mustExistFile(path string) error {
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("%s adalah directory, harus file", path)
	}
	return nil
}

func lookupEnvOrFile(key string, envFile string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	// EnvironmentFile pada unit systemd sifatnya optional (-), jadi file mungkin tidak ada.
	if _, err := os.Stat(envFile); err != nil {
		return ""
	}
	vals, err := godotenv.Read(envFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(vals[key])
}
