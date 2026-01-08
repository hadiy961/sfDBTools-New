// File : internal/app/backup/scheduler/run_job.go
// Deskripsi : Eksekusi job scheduler (dipanggil oleh systemd service)
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-02
// Last Modified : 2026-01-06
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
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/app/cleanup"
	cleanupmodel "sfdbtools/internal/app/cleanup/model"
	defaultVal "sfdbtools/internal/cli/defaults"
	appdeps "sfdbtools/internal/cli/deps"
	appconfig "sfdbtools/internal/services/config"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/encrypt"
)

func RunJob(ctx context.Context, deps *appdeps.Dependencies, jobName string) error {
	if strings.TrimSpace(jobName) == "" {
		return fmt.Errorf("--job wajib diisi")
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
		mode = consts.ModeSeparated
	}
	// Mapping kompatibilitas: allow user menulis "separated" / "combined".
	switch mode {
	case "separated":
		mode = consts.ModeSeparated
	case "combined":
		mode = consts.ModeCombined
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
		v, err := encrypt.ResolveEnvSecret(consts.ENV_SOURCE_PROFILE_KEY)
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
		opts.Ticket = fmt.Sprintf("SCHEDULED_%s", job.Name)
		deps.Logger.Warnf("ticket kosong, menggunakan default: %s", opts.Ticket)
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
			v, err := encrypt.ResolveEnvSecret(consts.ENV_BACKUP_ENCRYPTION_KEY)
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
		svc.HandleShutdown()
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
	if err := svc.ExecuteBackupCommand(runCtx, backupCfg); err != nil {
		if errors.Is(err, context.Canceled) || runCtx.Err() != nil {
			deps.Logger.Warn("Backup dibatalkan")
			return nil
		}
		return err
	}
	deps.Logger.Infof("Backup job '%s' selesai dalam %s", job.Name, time.Since(start))

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
		if err := cleanupSvc.ExecuteCleanupCommand(cleanupCfg); err != nil {
			return err
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
