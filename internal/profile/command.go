// File : internal/profile/command.go
// Deskripsi : Command execution functions untuk cmd layer
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2026-01-04

package profile

import (
	appdeps "sfDBTools/internal/deps"
	"sfDBTools/internal/parsing"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/ui"

	"github.com/spf13/cobra"
)

// =============================================================================
// Public API - Command Executors (dipanggil dari cmd layer)
// =============================================================================

// ExecuteProfile adalah unified function untuk menjalankan profile operations dengan mode apapun
func ExecuteProfile(cmd *cobra.Command, deps *appdeps.Dependencies, mode string) error {
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
func executeProfileWithConfig(cmd *cobra.Command, deps *appdeps.Dependencies, config types.ProfileEntryConfig) error {
	logger := deps.Logger
	if config.LogPrefix != "" {
		logger.Infof(consts.ProfileLogStartProcessWithPrefixFmt, config.LogPrefix, config.Mode)
	} else {
		logger.Infof(consts.ProfileLogStartProcessFmt, config.Mode)
	}

	// Parsing opsi berdasarkan mode
	var profileOptions interface{}
	var err error

	switch config.Mode {
	case consts.ProfileModeCreate:
		profileOptions, err = parsing.ParsingCreateProfile(cmd, logger)
	case consts.ProfileModeShow:
		profileOptions, err = parsing.ParsingShowProfile(cmd)
	case consts.ProfileModeEdit:
		profileOptions, err = parsing.ParsingEditProfile(cmd)
	case consts.ProfileModeDelete:
		profileOptions, err = parsing.ParsingDeleteProfile(cmd)
	default:
		return ErrInvalidProfileMode
	}

	if err != nil {
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
		consts.ProfileModeCreate: {
			HeaderTitle: consts.ProfileHeaderCreate,
			Mode:        consts.ProfileModeCreate,
			SuccessMsg:  consts.ProfileSuccessCreated,
			LogPrefix:   consts.ProfileLogPrefixCreate,
		},
		consts.ProfileModeShow: {
			HeaderTitle: consts.ProfileHeaderShow,
			Mode:        consts.ProfileModeShow,
			SuccessMsg:  "", // No success message for show
			LogPrefix:   consts.ProfileLogPrefixShow,
		},
		consts.ProfileModeEdit: {
			HeaderTitle: consts.ProfileHeaderEdit,
			Mode:        consts.ProfileModeEdit,
			SuccessMsg:  consts.ProfileSuccessUpdated,
			LogPrefix:   consts.ProfileLogPrefixEdit,
		},
		consts.ProfileModeDelete: {
			HeaderTitle: consts.ProfileHeaderDelete,
			Mode:        consts.ProfileModeDelete,
			SuccessMsg:  consts.ProfileSuccessDeleted,
			LogPrefix:   consts.ProfileLogPrefixDelete,
		},
	}

	config, ok := configs[mode]
	if !ok {
		return types.ProfileEntryConfig{}, ErrInvalidProfileMode
	}

	return config, nil
}
