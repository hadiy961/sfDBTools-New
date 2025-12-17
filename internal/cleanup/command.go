// File : internal/cleanup/command.go
// Deskripsi : Command execution functions untuk cmd layer
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-16

package cleanup

import (
	"sfDBTools/internal/types"
	"sfDBTools/internal/parsing"
	"sfDBTools/pkg/ui"

	"github.com/spf13/cobra"
)

// =============================================================================
// Public API - Command Executors (dipanggil dari cmd layer)
// =============================================================================

// ExecuteCleanup adalah unified function untuk menjalankan cleanup dengan mode apapun
func ExecuteCleanup(cmd *cobra.Command, deps *types.Dependencies, mode string) error {
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
func executeCleanupWithConfig(cmd *cobra.Command, deps *types.Dependencies, config types.CleanupEntryConfig) error {
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

	// Tampilkan header jika ada
	if config.HeaderTitle != "" {
		ui.Headers(config.HeaderTitle)
	}

	// Execute cleanup command
	if err := svc.ExecuteCleanupCommand(config); err != nil {
		return err
	}

	// Print success message jika ada
	if config.SuccessMsg != "" {
		ui.PrintSuccess(config.SuccessMsg)
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
			DryRun:      false,
		},
		"dry-run": {
			HeaderTitle: "Cleanup Preview (Dry Run)",
			Mode:        "dry-run",
			ShowOptions: false,
			SuccessMsg:  "✓ Cleanup preview selesai",
			LogPrefix:   "cleanup-dryrun",
			DryRun:      true,
		},
		"pattern": {
			HeaderTitle: "Cleanup By Pattern",
			Mode:        "pattern",
			ShowOptions: false,
			SuccessMsg:  "✓ Cleanup by pattern selesai",
			LogPrefix:   "cleanup-pattern",
			DryRun:      false,
		},
	}

	config, ok := configs[mode]
	if !ok {
		return types.CleanupEntryConfig{}, ErrInvalidCleanupMode
	}

	return config, nil
}
