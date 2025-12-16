// File : internal/profile/command.go
// Deskripsi : Command execution functions untuk cmd layer
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-16

package profile

import (
	"sfDBTools/internal/types"
	"sfDBTools/pkg/parsing"
	"sfDBTools/pkg/ui"

	"github.com/spf13/cobra"
)

// =============================================================================
// Public API - Command Executors (dipanggil dari cmd layer)
// =============================================================================

// ExecuteProfile adalah unified function untuk menjalankan profile operations dengan mode apapun
func ExecuteProfile(cmd *cobra.Command, deps *types.Dependencies, mode string) error {
	// Dapatkan konfigurasi execution berdasarkan mode
	config, err := GetExecutionConfig(mode)
	if err != nil {
		return err
	}

	return executeProfileWithConfig(cmd, deps, config)
}

// =============================================================================
// Internal Helpers
// =============================================================================

// executeProfileWithConfig adalah helper function yang menjalankan profile dengan configuration
func executeProfileWithConfig(cmd *cobra.Command, deps *types.Dependencies, config types.ProfileEntryConfig) error {
	logger := deps.Logger
	logger.Info("Memulai proses profile - " + config.Mode)

	// Parsing opsi berdasarkan mode
	var profileOptions interface{}
	var err error

	switch config.Mode {
	case "create":
		profileOptions, err = parsing.ParsingCreateProfile(cmd, logger)
	case "show":
		profileOptions, err = parsing.ParsingShowProfile(cmd)
	case "edit":
		profileOptions, err = parsing.ParsingEditProfile(cmd)
	case "delete":
		profileOptions, err = parsing.ParsingDeleteProfile(cmd)
	default:
		return ErrInvalidProfileMode
	}

	if err != nil {
		logger.Error("gagal parsing opsi: " + err.Error())
		return err
	}

	// Inisialisasi service profile
	svc := NewProfileService(deps.Config, logger, profileOptions)

	// Tampilkan header jika ada
	if config.HeaderTitle != "" {
		ui.Headers(config.HeaderTitle)
	}

	// Execute profile command
	if err := svc.ExecuteProfileCommand(config); err != nil {
		return err
	}

	// Print success message jika ada
	if config.SuccessMsg != "" {
		ui.PrintSuccess(config.SuccessMsg)
		logger.Info(config.SuccessMsg)
	}

	return nil
}

// GetExecutionConfig mengembalikan konfigurasi untuk mode profile tertentu
func GetExecutionConfig(mode string) (types.ProfileEntryConfig, error) {
	configs := map[string]types.ProfileEntryConfig{
		"create": {
			HeaderTitle: "Create Database Profile",
			Mode:        "create",
			ShowOptions: false,
			SuccessMsg:  "✓ Profile berhasil dibuat",
			LogPrefix:   "profile-create",
		},
		"show": {
			HeaderTitle: "Show Database Profile",
			Mode:        "show",
			ShowOptions: false,
			SuccessMsg:  "", // No success message for show
			LogPrefix:   "profile-show",
		},
		"edit": {
			HeaderTitle: "Edit Database Profile",
			Mode:        "edit",
			ShowOptions: false,
			SuccessMsg:  "✓ Profile berhasil diupdate",
			LogPrefix:   "profile-edit",
		},
		"delete": {
			HeaderTitle: "Delete Database Profile",
			Mode:        "delete",
			ShowOptions: false,
			SuccessMsg:  "✓ Profile berhasil dihapus",
			LogPrefix:   "profile-delete",
		},
	}

	config, ok := configs[mode]
	if !ok {
		return types.ProfileEntryConfig{}, ErrInvalidProfileMode
	}

	return config, nil
}
