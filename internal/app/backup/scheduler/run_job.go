// File : internal/app/backup/scheduler/run_job.go
// Deskripsi : Eksekusi job scheduler (dipanggil oleh systemd service)
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-02
// Last Modified : 2026-01-20
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"sfdbtools/internal/app/backup"
	backuppath "sfdbtools/internal/app/backup/helpers/path"
	"sfdbtools/internal/app/backup/model"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/app/cleanup"
	cleanupmodel "sfdbtools/internal/app/cleanup/model"
	defaultVal "sfdbtools/internal/cli/defaults"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/crypto"
	appconfig "sfdbtools/internal/services/config"
	"sfdbtools/internal/shared/consts"
)

func RunJob(ctx context.Context, deps *appdeps.Dependencies, jobName string) error {
	if strings.TrimSpace(jobName) == "" {
		return fmt.Errorf("RunScheduledJob: %w", model.ErrJobNameRequired)
	}
	jobs, err := getTargetJobs(deps, jobName, false)
	if err != nil {
		return err
	}
	job := jobs[0]
	if !job.Enabled {
		return fmt.Errorf("job '%s' tidak enabled di config", jobName)
	}

	mode := strings.TrimSpace(job.Mode)
	if mode == "" {
		return fmt.Errorf("job '%s': mode wajib diisi", job.Name)
	}

	// Normalize + validate mode
	switch strings.ToLower(mode) {
	case "separated":
		mode = consts.ModeSeparated
	case "combined":
		mode = consts.ModeCombined
	case "all":
		mode = consts.ModeAll
	case "single":
		mode = consts.ModeSingle
	case "primary":
		mode = consts.ModePrimary
	case "secondary":
		mode = consts.ModeSecondary
	default:
		return fmt.Errorf("job '%s': mode '%s' tidak valid (valid: separated|combined|all|single|primary|secondary)", job.Name, mode)
	}

	opts := defaultVal.DefaultBackupOptions(mode)
	opts.NonInteractive = true

	// Profile
	profilePath := strings.TrimSpace(job.Profile)
	if profilePath == "" {
		profilePath = strings.TrimSpace(os.Getenv(consts.ENV_SOURCE_PROFILE))
	}
	if profilePath == "" {
		return fmt.Errorf("profile wajib untuk scheduler: set backup.scheduler.jobs[].profile atau env %s", consts.ENV_SOURCE_PROFILE)
	}
	opts.Profile.Path = profilePath
	// profile-key boleh dari env / config; untuk scheduler kita fail-fast jika kosong.
	if strings.TrimSpace(opts.Profile.EncryptionKey) == "" {
		v, err := crypto.ResolveEnvSecret(consts.ENV_SOURCE_PROFILE_KEY)
		if err != nil {
			return err
		}
		opts.Profile.EncryptionKey = strings.TrimSpace(v)
	}
	if strings.TrimSpace(opts.Profile.EncryptionKey) == "" {
		return fmt.Errorf("profile-key wajib untuk scheduler: set env %s", consts.ENV_SOURCE_PROFILE_KEY)
	}

	// Ticket
	opts.Ticket = strings.TrimSpace(job.Ticket)
	if opts.Ticket == "" {
		return fmt.Errorf("job '%s': ticket wajib diisi (untuk audit trail)", job.Name)
	}

	// Filters: scheduler ini fokus untuk backup filter/separated via include file.
	if mode == consts.ModeSeparated || mode == consts.ModeCombined {
		includeFile := strings.TrimSpace(job.IncludeFile)
		if includeFile == "" {
			return fmt.Errorf("include_file wajib untuk job '%s' (mode %s)", job.Name, mode)
		}
		opts.Filter.IncludeFile = includeFile
	}

	// Output directory: override base_directory per job, struktur mengikuti config global.
	baseDir := strings.TrimSpace(job.Output.BaseDirectory)
	if baseDir == "" {
		baseDir = deps.Config.Backup.Output.BaseDirectory
	}

	outputDir, err := backuppath.GenerateBackupDirectory(
		baseDir,
		deps.Config.Backup.Output.Structure.Pattern,
		"",
	)
	if err != nil {
		return fmt.Errorf("gagal membuat output directory untuk job '%s': %w", job.Name, err)
	}
	opts.OutputDir = outputDir

	// Encryption: wajib ada key jika enabled.
	if opts.Encryption.Enabled {
		if strings.TrimSpace(opts.Encryption.Key) == "" {
			// Coba dari env
			v, err := crypto.ResolveEnvSecret(consts.ENV_BACKUP_ENCRYPTION_KEY)
			if err != nil {
				return err
			}
			opts.Encryption.Key = strings.TrimSpace(v)
		}
		if strings.TrimSpace(opts.Encryption.Key) == "" {
			return fmt.Errorf("backup encryption aktif tapi key kosong: set config backup.encryption.key atau env %s (atau disable encryption)", consts.ENV_BACKUP_ENCRYPTION_KEY)
		}
	}

	// Inisialisasi service backup
	svc := backup.NewBackupService(deps.Logger, deps.Config, &opts)

	// Buat execution state untuk tracking (dibuat early agar bisa diakses oleh signal handler)
	execState := backup.NewExecutionState()

	// Setup context + cancellation
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	svc.SetCancelFunc(cancel)

	// Setup signal handler (SIGTERM dari systemd)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	go func() {
		sig := <-sigChan
		deps.Logger.Warnf("Menerima signal %v, menghentikan backup job '%s'...", sig, job.Name)
		svc.HandleShutdown(execState)
		cancel()
	}()

	// Header + config execution
	execCfg, err := backup.GetExecutionConfig(mode)
	if err != nil {
		return err
	}
	backupCfg := types_backup.BackupEntryConfig{
		HeaderTitle:    fmt.Sprintf("Scheduled Backup (%s)", job.Name),
		NonInteractive: true,
		SuccessMsg:     execCfg.SuccessMsg,
		LogPrefix:      fmt.Sprintf("schedule-%s", job.Name),
		BackupMode:     execCfg.Mode,
	}

	start := time.Now()
	if err := svc.ExecuteBackupCommand(runCtx, execState, backupCfg); err != nil {
		duration := time.Since(start)

		// Handle cancellation (SIGTERM/SIGINT) - ini bukan error
		if errors.Is(err, context.Canceled) || runCtx.Err() != nil {
			deps.Logger.Warnf("Backup job '%s' dibatalkan (duration=%s)", job.Name, duration)
			return nil
		}

		// Wrap error dengan scheduler context untuk troubleshooting
		wrappedErr := wrapSchedulerError(err, job.Name, mode, duration)
		deps.Logger.Error(wrappedErr.Error())

		return wrappedErr
	}
	deps.Logger.Infof("✓ Backup job '%s' selesai dalam %s", job.Name, time.Since(start))

	// Cleanup (retention) per job
	if job.Cleanup.Enabled && job.Cleanup.RetentionDays > 0 {
		cfgCopy := cloneConfig(deps.Config)
		cfgCopy.Backup.Output.BaseDirectory = baseDir
		cfgCopy.Backup.Cleanup.Enabled = true
		cfgCopy.Backup.Cleanup.Days = job.Cleanup.RetentionDays

		cleanupSvc := cleanup.NewCleanupService(&cfgCopy, deps.Logger, cleanupmodel.CleanupOptions{
			Enabled:    true,
			Days:       job.Cleanup.RetentionDays,
			Background: false,
		})

		cleanupCfg, err := cleanup.GetExecutionConfig("run")
		if err != nil {
			return err
		}
		cleanupCfg.LogPrefix = fmt.Sprintf("schedule-%s-cleanup", job.Name)

		deps.Logger.Infof("Menjalankan cleanup job '%s' (retention=%d hari)", job.Name, job.Cleanup.RetentionDays)
		cleanupStart := time.Now()
		if err := cleanupSvc.ExecuteCleanupCommand(cleanupCfg); err != nil {
			// Cleanup error tidak boleh membatalkan backup yang sudah berhasil
			// Log as warning dan lanjutkan (cleanup akan di-retry di next run)
			deps.Logger.Warnf("Cleanup job '%s' gagal (duration=%s): %v",
				job.Name, time.Since(cleanupStart), err)
			deps.Logger.Warn("Backup berhasil, tapi cleanup gagal. Cleanup akan di-retry di run berikutnya.")
			// Tidak return error agar systemd tidak menganggap job gagal
		} else {
			deps.Logger.Infof("✓ Cleanup job '%s' selesai dalam %s", job.Name, time.Since(cleanupStart))
		}
	}

	return nil
}

func cloneConfig(cfg *appconfig.Config) appconfig.Config {
	if cfg == nil {
		return appconfig.Config{}
	}
	return *cfg
}

// IsTransientError mengklasifikasikan error sebagai transient (bisa di-retry) atau permanent.
// Transient errors adalah error yang kemungkinan besar akan berhasil jika di-retry setelah beberapa waktu.
// Function ini di-export untuk digunakan di cmd layer untuk semantic exit codes.
func IsTransientError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())

	// Pattern untuk transient errors (network, resource, temporary issues)
	transientPatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"timed out",
		"too many connections",
		"no space left",
		"disk full",
		"temporary failure",
		"resource temporarily unavailable",
		"i/o timeout",
		"broken pipe",
		"network is unreachable",
		"host is unreachable",
	}

	for _, pattern := range transientPatterns {
		if strings.Contains(msg, pattern) {
			return true
		}
	}

	return false
}

// wrapSchedulerError menambahkan context scheduler ke error message untuk troubleshooting.
func wrapSchedulerError(err error, jobName, mode string, duration time.Duration) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("scheduled job '%s' failed (mode=%s, duration=%s): %w",
		jobName, mode, duration, err)
}
