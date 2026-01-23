// File : internal/app/cleanup/command.go
// Deskripsi : Command execution functions untuk cmd layer
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 6 Januari 2026
package cleanup

import (
	cleanupmodel "sfdbtools/internal/app/cleanup/model"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/parsing"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/ui/print"

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
func executeCleanupWithConfig(cmd *cobra.Command, deps *appdeps.Dependencies, config cleanupmodel.CleanupEntryConfig) error {
	logger := deps.Logger
	logger.Info("Memulai proses cleanup - " + config.Mode)

	// Parsing opsi dari flags
	parsedOpts, err := parsing.ParsingCleanupOptions(cmd)
	if err != nil {
		logger.Error("gagal parsing opsi: " + err.Error())
		return err
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
	if config.HeaderTitle != "" && !runtimecfg.IsQuiet() {
		print.PrintAppHeader(config.HeaderTitle)
	}

	// Execute cleanup command
	if err := svc.ExecuteCleanupCommand(config); err != nil {
		return err
	}

	// Print success message jika ada (skip stdout saat quiet/daemon)
	if config.SuccessMsg != "" {
		if !runtimecfg.IsQuiet() {
			print.PrintSuccess(config.SuccessMsg)
		}
		logger.Info(config.SuccessMsg)
	}

	return nil
}

// GetExecutionConfig mengembalikan konfigurasi untuk mode cleanup tertentu
func GetExecutionConfig(mode string) (cleanupmodel.CleanupEntryConfig, error) {
	configs := map[string]cleanupmodel.CleanupEntryConfig{
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
		return cleanupmodel.CleanupEntryConfig{}, ErrInvalidCleanupMode
	}

	return config, nil
}
