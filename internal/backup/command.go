// File : internal/backup/command.go
// Deskripsi : Command execution functions untuk cmd layer
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package backup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sfDBTools/internal/backup/display"
	"sfDBTools/internal/parsing"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
	"syscall"

	"github.com/spf13/cobra"
)

// =============================================================================
// Public API - Command Executors (dipanggil dari cmd layer)
// =============================================================================

// ExecuteBackup adalah unified function untuk menjalankan backup dengan mode apapun
// Menggantikan 5 fungsi Execute* yang duplikat (Single, Separated, Combined, Primary, Secondary)
func ExecuteBackup(cmd *cobra.Command, deps *types.Dependencies, mode string) error {
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
func (s *Service) ExecuteBackupCommand(ctx context.Context, config types_backup.BackupEntryConfig) error {
	// Setup session (koneksi database source)
	sourceClient, dbFiltered, err := s.PrepareBackupSession(ctx, config.HeaderTitle, config.Force)
	if err != nil {
		return err
	}

	// Cleanup function untuk close semua connections
	defer func() {
		if sourceClient != nil {
			sourceClient.Close()
		}
	}()

	// Lakukan backup
	result, err := s.ExecuteBackup(ctx, sourceClient, dbFiltered, config.BackupMode)
	if err != nil {
		return err
	}

	// Tampilkan hasil
	display.NewResultDisplayer(
		result,
		s.BackupDBOptions.Compression.Enabled,
		s.BackupDBOptions.Compression.Type,
		s.BackupDBOptions.Encryption.Enabled,
	).Display()

	// Print success message jika ada
	if config.SuccessMsg != "" {
		ui.PrintSuccess(config.SuccessMsg)
		s.Log.Info(config.SuccessMsg)
	}
	return nil
}

// =============================================================================
// Internal Helpers
// =============================================================================

// executeBackupWithConfig adalah helper function yang menjalankan backup dengan configuration
func executeBackupWithConfig(cmd *cobra.Command, deps *types.Dependencies, config types_backup.ExecutionConfig) error {
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
		fmt.Println()
		logger.Warnf("Menerima signal %v, menghentikan backup... (Tekan sekali lagi untuk force exit)", sig)
		svc.HandleShutdown()
		cancel()

		<-sigChan
		fmt.Println()
		logger.Warn("Menerima signal kedua, memaksa berhenti (force exit)...")
		os.Exit(1)
	}()

	// BackupEntryConfig menyimpan konfigurasi untuk proses backup
	backupConfig := types_backup.BackupEntryConfig{
		HeaderTitle: config.HeaderTitle,
		Force:       parsedOpts.Force,
		SuccessMsg:  config.SuccessMsg,
		LogPrefix:   config.LogPrefix,
		BackupMode:  config.Mode,
	}

	if err := svc.ExecuteBackupCommand(ctx, backupConfig); err != nil {
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
