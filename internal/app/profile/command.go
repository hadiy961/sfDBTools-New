// File : internal/profile/command.go
// Deskripsi : Command execution functions untuk cmd layer
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 5 Januari 2026
package profile

import (
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/parsing"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"

	profilemodel "sfdbtools/internal/app/profile/model"

	"github.com/spf13/cobra"
)

var profileExecutionConfigs = map[string]profilemodel.ProfileEntryConfig{
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
func executeProfileWithConfig(cmd *cobra.Command, deps *appdeps.Dependencies, config profilemodel.ProfileEntryConfig) error {
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
		print.PrintAppHeader(config.HeaderTitle)
	}

	// Execute profile command
	if err := svc.ExecuteProfileCommand(config); err != nil {
		return err
	}

	// Print success message jika ada
	if config.SuccessMsg != "" {
		print.PrintSuccess(config.SuccessMsg)
		logger.Info(config.SuccessMsg)
	}

	return nil
}

// GetExecutionConfig mengembalikan konfigurasi untuk mode profile tertentu
func GetExecutionConfig(mode string) (profilemodel.ProfileEntryConfig, error) {
	config, ok := profileExecutionConfigs[mode]
	if !ok {
		return profilemodel.ProfileEntryConfig{}, ErrInvalidProfileMode
	}

	return config, nil
}
