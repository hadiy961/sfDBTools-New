// File : internal/backup/command.go
// Deskripsi : Command execution functions untuk cmd layer
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2026-01-06
package backup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sfdbtools/internal/app/backup/display"
	"sfdbtools/internal/app/backup/helpers/path"
	"sfdbtools/internal/app/backup/model"
	"sfdbtools/internal/app/backup/model/types_backup"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/parsing"
	"sfdbtools/internal/services/scheduler"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/style"
	"sfdbtools/internal/ui/text"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// =============================================================================
// Public API - Command Executors (dipanggil dari cmd layer)
// =============================================================================

// ExecuteBackup adalah unified function untuk menjalankan backup dengan mode apapun
// Menggantikan 5 fungsi Execute* yang duplikat (Single, Separated, Combined, Primary, Secondary)
func ExecuteBackup(cmd *cobra.Command, deps *appdeps.Dependencies, mode string) error {
	// Dapatkan konfigurasi execution berdasarkan mode
	config, err := GetExecutionConfig(mode)
	if err != nil {
		return err
	}

	return executeBackupWithConfig(cmd, deps, config)
}

// =============================================================================
// Entry Point Execution
// =============================================================================

// ExecuteBackupCommand adalah unified entry point untuk semua jenis backup
func (s *Service) ExecuteBackupCommand(ctx context.Context, state *BackupExecutionState, config types_backup.BackupEntryConfig) error {
	// Setup session (koneksi database source)
	// RESOURCE OWNERSHIP: PrepareBackupSession transfer ownership ke caller.
	// Jika PrepareBackupSession return error, client sudah di-close otomatis (tidak perlu cleanup).
	// Jika return success, WAJIB close client sebelum function exit.
	sourceClient, dbFiltered, err := s.PrepareBackupSession(ctx, state, config.HeaderTitle, config.NonInteractive)
	if err != nil {
		// Client sudah di-close di PrepareBackupSession jika error
		return err
	}

	// CRITICAL: Register cleanup IMMEDIATELY setelah resource acquisition berhasil.
	// Pattern ini memastikan client selalu di-close bahkan jika terjadi panic/error.
	defer func() {
		if sourceClient != nil {
			sourceClient.Close()
			sourceClient = nil // prevent double-close
		}
	}()

	// Lakukan backup (returns result, state, error)
	result, _, err := s.ExecuteBackup(ctx, state, sourceClient, dbFiltered, config.BackupMode)
	if err != nil {
		return err
	}

	// Tampilkan hasil (skip output interaktif saat quiet/daemon)
	if !runtimecfg.IsQuiet() && !runtimecfg.IsDaemon() {
		display.NewResultDisplayer(
			result,
			s.BackupDBOptions.Compression.Enabled,
			s.BackupDBOptions.Compression.Type,
			s.BackupDBOptions.Encryption.Enabled,
		).Display()
	}

	// Print success message jika ada
	if config.SuccessMsg != "" {
		if !runtimecfg.IsQuiet() && !runtimecfg.IsDaemon() {
			print.PrintSuccess(config.SuccessMsg)
		}
		s.Log.Info(config.SuccessMsg)
	}
	return nil
}

// =============================================================================
// Internal Helpers
// =============================================================================
// executeBackupWithConfig adalah helper function yang menjalankan backup dengan configuration
func executeBackupWithConfig(cmd *cobra.Command, deps *appdeps.Dependencies, config types_backup.ExecutionConfig) error {
	logger := deps.Logger
	logger.Info("Memulai proses backup database - " + config.Mode)

	// Jika diminta background dan bukan proses daemon, jalankan ulang command via systemd-run.
	bg := false
	if cmd.Flags().Lookup("background") != nil {
		v, err := cmd.Flags().GetBool("background")
		if err == nil {
			bg = v
		}
	}
	if bg && !runtimecfg.IsDaemon() {
		// Backup background harus non-interaktif agar tidak hang tanpa TTY.
		nonInteractive := runtimecfg.IsQuiet() || runtimecfg.IsDaemon()
		if !nonInteractive {
			return fmt.Errorf("RunBackupCommand: %w", model.ErrBackgroundModeRequiresQuiet)
		}

		// Parse profile path untuk generate unique lock file
		// Setiap profile mendapat lock file sendiri untuk prevent concurrent backup conflicts
		profilePath, _ := cmd.Flags().GetString("profile")
		lockFilePath := path.GenerateForProfile(profilePath)

		wd, _ := os.Getwd()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		res, runErr := scheduler.SpawnSelfInBackground(ctx, scheduler.SpawnSelfOptions{
			UnitPrefix:    "sfdbtools-backup",
			Mode:          scheduler.RunModeAuto,
			EnvFile:       "/etc/sfDBTools/.env",
			WorkDir:       wd,
			Collect:       true,
			NoAskPass:     true,
			WrapWithFlock: true,
			FlockPath:     lockFilePath,
		})
		if runErr != nil {
			return runErr
		}

		print.PrintHeader("DB BACKUP - BACKGROUND MODE")
		print.PrintSuccess("Background backup dimulai via systemd")
		print.PrintInfo(fmt.Sprintf("Unit: %s", text.Color(res.UnitName, style.ColorCyan)))
		if res.Mode == scheduler.RunModeUser {
			print.PrintInfo(fmt.Sprintf("Status: systemctl --user status %s", res.UnitName))
			print.PrintInfo(fmt.Sprintf("Logs: journalctl --user -u %s -f", res.UnitName))
		} else {
			print.PrintInfo(fmt.Sprintf("Status: sudo systemctl status %s", res.UnitName))
			print.PrintInfo(fmt.Sprintf("Logs: sudo journalctl -u %s -f", res.UnitName))
		}
		return nil
	}

	// Parsing opsi berdasarkan mode
	parsedOpts, err := parsing.ParsingBackupOptions(cmd, config.Mode)
	if err != nil {
		logger.Error("gagal parsing opsi: " + err.Error())
		return err
	}

	// Inisialisasi service backup
	svc := NewBackupService(logger, deps.Config, &parsedOpts)

	// Buat execution state untuk tracking (dibuat early agar bisa diakses oleh signal handler)
	execState := NewExecutionState()

	// Setup context dengan cancellation untuk graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set cancel function ke service untuk graceful shutdown
	svc.SetCancelFunc(cancel)

	// Setup signal handler untuk CTRL+C (SIGINT) dan SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Goroutine untuk menangani signal
	go func() {
		sig := <-sigChan
		if !runtimecfg.IsQuiet() && !runtimecfg.IsDaemon() {
			print.Println()
		}
		logger.Warnf("Menerima signal %v, menghentikan backup... (Tekan sekali lagi untuk force exit)", sig)
		svc.HandleShutdown(execState)
		cancel()

		<-sigChan
		if !runtimecfg.IsQuiet() && !runtimecfg.IsDaemon() {
			print.Println()
		}
		logger.Warn("Menerima signal kedua, memaksa berhenti (force exit)...")

		// Issue #58: Lakukan critical cleanup sebelum os.Exit()
		// os.Exit() bypasses all deferred cleanup, menyebabkan:
		// - Stale lock files (deadlock pada backup selanjutnya)
		// - Partial backup files tidak ter-cleanup (disk space waste)
		// - File handles tidak ter-flush (corrupted files)
		// Critical cleanup memastikan resources kritis di-cleanup sebelum force exit
		criticalCleanup(execState, logger)

		os.Exit(1)
	}()

	// BackupEntryConfig menyimpan konfigurasi untuk proses backup
	backupConfig := types_backup.BackupEntryConfig{
		HeaderTitle:    config.HeaderTitle,
		NonInteractive: parsedOpts.NonInteractive,
		SuccessMsg:     config.SuccessMsg,
		LogPrefix:      config.LogPrefix,
		BackupMode:     config.Mode,
	}

	if err := svc.ExecuteBackupCommand(ctx, execState, backupConfig); err != nil {
		if errors.Is(err, validation.ErrUserCancelled) {
			logger.Warn("Proses dibatalkan oleh pengguna.")
			return nil
		}
		if ctx.Err() != nil || errors.Is(err, context.Canceled) {
			logger.Warn("Proses backup dibatalkan.")
			return nil
		}
		logger.Error("backup gagal (" + config.Mode + "): " + err.Error())
		return err
	}

	return nil
}
