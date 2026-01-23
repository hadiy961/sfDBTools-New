// File : internal/backup/command.go
// Deskripsi : Command execution functions untuk cmd layer
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2026-01-23
package backup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sfdbtools/internal/app/backup/display"
	"sfdbtools/internal/app/backup/model/types_backup"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/parsing"
	resolver "sfdbtools/internal/cli/resolver"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"strings"
	"syscall"

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

	// Tampilkan hasil (skip output interaktif saat quiet)
	if !runtimecfg.IsQuiet() {
		display.NewResultDisplayer(
			result,
			s.BackupDBOptions.Compression.Enabled,
			s.BackupDBOptions.Compression.Type,
			s.BackupDBOptions.Encryption.Enabled,
		).Display()
	}

	// Print success message jika ada
	if config.SuccessMsg != "" {
		if !runtimecfg.IsQuiet() {
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
		if !runtimecfg.IsQuiet() {
			print.Println()
		}
		logger.Warnf("Menerima signal %v, menghentikan backup... (Tekan sekali lagi untuk force exit)", sig)
		svc.HandleShutdown(execState)
		cancel()

		<-sigChan
		if !runtimecfg.IsQuiet() {
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

func validateBackupNonInteractivePreflight(cmd *cobra.Command, deps *appdeps.Dependencies, mode string) error {
	// Preflight ini mengikuti aturan parsing non-interaktif (ParsingBackupOptions).
	// Tujuannya agar background mode tidak membuat transient unit yang langsung gagal.

	ticket := strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "ticket", ""))
	if ticket == "" {
		return fmt.Errorf("ticket wajib diisi pada mode background: gunakan --ticket")
	}

	profilePath := strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "profile", consts.ENV_SOURCE_PROFILE))
	if profilePath == "" {
		return fmt.Errorf("profile wajib diisi pada mode background: gunakan --profile atau env %s", consts.ENV_SOURCE_PROFILE)
	}
	profileKey, err := resolver.GetSecretStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY)
	if err != nil {
		return err
	}
	if strings.TrimSpace(profileKey) == "" {
		return fmt.Errorf("profile-key wajib diisi pada mode background: gunakan --profile-key atau env %s", consts.ENV_SOURCE_PROFILE_KEY)
	}

	// Encryption key requirement (jika enkripsi aktif dan tidak di-skip).
	skipEncrypt := resolver.GetBoolFlagOrEnv(cmd, "skip-encrypt", "")
	encryptionEnabledByDefault := false
	if deps != nil && deps.Config != nil {
		encryptionEnabledByDefault = deps.Config.Backup.Encryption.Enabled || strings.TrimSpace(deps.Config.Backup.Encryption.Key) != ""
	}
	if encryptionEnabledByDefault && !skipEncrypt {
		backupKey, err := resolver.GetSecretStringFlagOrEnv(cmd, "backup-key", consts.ENV_BACKUP_ENCRYPTION_KEY)
		if err != nil {
			return err
		}
		if strings.TrimSpace(backupKey) == "" {
			return fmt.Errorf("backup-key wajib diisi saat enkripsi aktif pada mode background: gunakan --backup-key atau env %s (atau set --skip-encrypt)", consts.ENV_BACKUP_ENCRYPTION_KEY)
		}
	}

	// Mode-specific requirements.
	if mode == consts.ModeSingle {
		dbName := strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "database", ""))
		if dbName == "" {
			return fmt.Errorf("database wajib diisi pada mode backup single saat background: gunakan --database")
		}
	}
	// Mode primary: saat background WAJIB pakai client-code untuk menghindari backup massal tanpa sengaja.
	if mode == consts.ModePrimary {
		clientCode := strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "client-code", ""))
		if clientCode == "" {
			return fmt.Errorf("mode backup primary saat background wajib mengisi client-code: gunakan --client-code")
		}
	}
	// Filter command: butuh include list saat background/non-interaktif.
	if cmd.Use == "filter" || cmd.Name() == "filter" {
		dbFile := strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "db-file", ""))
		db := resolver.GetStringArrayFlagOrEnv(cmd, "db", "")
		if len(db) == 0 && dbFile == "" {
			return fmt.Errorf("mode backup filter background membutuhkan include list: gunakan --db atau --db-file")
		}
	}

	return nil
}
