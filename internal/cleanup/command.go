// File : internal/cleanup/command.go
// Deskripsi : Command execution functions untuk cmd layer
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified :  2026-01-05
package cleanup

import (
	"context"
	"fmt"
	"os"
	"time"

	appdeps "sfDBTools/internal/deps"
	"sfDBTools/internal/parsing"
	"sfDBTools/internal/services/scheduler"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/runtimecfg"
	"sfDBTools/pkg/ui"

	"github.com/spf13/cobra"
)

// =============================================================================
// Public API - Command Executors (dipanggil dari cmd layer)
// =============================================================================

// ExecuteCleanup adalah unified function untuk menjalankan cleanup dengan mode apapun
func ExecuteCleanup(cmd *cobra.Command, deps *appdeps.Dependencies, mode string) error {
	// Dapatkan konfigurasi execution berdasarkan mode
	config, err := GetExecutionConfig(mode)
	if err != nil {
		return err
	}

	return executeCleanupWithConfig(cmd, deps, config)
}

// =============================================================================
// Internal Helpers
// =============================================================================

// executeCleanupWithConfig adalah helper function yang menjalankan cleanup dengan configuration
func executeCleanupWithConfig(cmd *cobra.Command, deps *appdeps.Dependencies, config types.CleanupEntryConfig) error {
	logger := deps.Logger
	logger.Info("Memulai proses cleanup - " + config.Mode)

	// Parsing opsi dari flags
	parsedOpts, err := parsing.ParsingCleanupOptions(cmd)
	if err != nil {
		logger.Error("gagal parsing opsi: " + err.Error())
		return err
	}

	// Konsisten dengan backup/dbscan: jika diminta background dan bukan proses daemon,
	// jalankan ulang command ini via systemd-run (transient unit) dengan flag --daemon.
	if parsedOpts.Background && !runtimecfg.IsDaemon() {
		wd, _ := os.Getwd()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		res, runErr := schedulerutil.SpawnSelfInBackground(ctx, schedulerutil.SpawnSelfOptions{
			UnitPrefix:    "sfdbtools-cleanup",
			Mode:          schedulerutil.RunModeAuto,
			EnvFile:       "/etc/sfDBTools/.env",
			WorkDir:       wd,
			Collect:       true,
			NoAskPass:     true,
			WrapWithFlock: false,
		})
		if runErr != nil {
			return runErr
		}
		ui.PrintHeader("CLEANUP - BACKGROUND MODE")
		ui.PrintSuccess("Background cleanup dimulai via systemd")
		ui.PrintInfo(fmt.Sprintf("Unit: %s", ui.ColorText(res.UnitName, consts.UIColorCyan)))
		if res.Mode == schedulerutil.RunModeUser {
			ui.PrintInfo(fmt.Sprintf("Status: systemctl --user status %s", res.UnitName))
			ui.PrintInfo(fmt.Sprintf("Logs: journalctl --user -u %s -f", res.UnitName))
		} else {
			ui.PrintInfo(fmt.Sprintf("Status: sudo systemctl status %s", res.UnitName))
			ui.PrintInfo(fmt.Sprintf("Logs: sudo journalctl -u %s -f", res.UnitName))
		}
		return nil
	}

	// Inisialisasi service cleanup
	svc := NewCleanupService(deps.Config, logger, parsedOpts)

	// Jika dry-run diminta via flag, gunakan judul & pesan yang konsisten dengan mode preview.
	if parsedOpts.DryRun {
		config.HeaderTitle = "Cleanup Preview (Dry Run)"
		config.SuccessMsg = "✓ Cleanup preview selesai"
		config.LogPrefix = "cleanup-dryrun"
	}

	// Tampilkan header jika ada (skip saat quiet/daemon)
	if config.HeaderTitle != "" && !runtimecfg.IsQuiet() && !runtimecfg.IsDaemon() {
		ui.Headers(config.HeaderTitle)
	}

	// Execute cleanup command
	if err := svc.ExecuteCleanupCommand(config); err != nil {
		return err
	}

	// Print success message jika ada (skip stdout saat quiet/daemon)
	if config.SuccessMsg != "" {
		if !runtimecfg.IsQuiet() && !runtimecfg.IsDaemon() {
			ui.PrintSuccess(config.SuccessMsg)
		}
		logger.Info(config.SuccessMsg)
	}

	return nil
}

// GetExecutionConfig mengembalikan konfigurasi untuk mode cleanup tertentu
func GetExecutionConfig(mode string) (types.CleanupEntryConfig, error) {
	configs := map[string]types.CleanupEntryConfig{
		"run": {
			HeaderTitle: "Cleanup Old Backup Files",
			Mode:        "run",
			ShowOptions: false,
			SuccessMsg:  "✓ Cleanup backup files selesai",
			LogPrefix:   "cleanup-run",
		},
		"pattern": {
			HeaderTitle: "Cleanup By Pattern",
			Mode:        "pattern",
			ShowOptions: false,
			SuccessMsg:  "✓ Cleanup by pattern selesai",
			LogPrefix:   "cleanup-pattern",
		},
	}

	config, ok := configs[mode]
	if !ok {
		return types.CleanupEntryConfig{}, ErrInvalidCleanupMode
	}

	return config, nil
}
