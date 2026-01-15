// File : internal/profile/command.go
// Deskripsi : Command execution functions untuk cmd layer
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 15 Januari 2026
package profile

import (
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/parsing"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"

	profileerrors "sfdbtools/internal/app/profile/errors"
	profilemodel "sfdbtools/internal/app/profile/model"

	"github.com/spf13/cobra"
)

// ProfileCommand merepresentasikan satu operasi profile di command layer.
// Tujuan: meminimalkan coupling command layer terhadap implementasi Service.
type ProfileCommand interface {
	Execute() error
}

var profileExecutionConfigs = map[string]profilemodel.ProfileEntryConfig{
	consts.ProfileModeCreate: {
		HeaderTitle: consts.ProfileUIHeaderCreate,
		Mode:        consts.ProfileModeCreate,
		SuccessMsg:  consts.ProfileSuccessCreated,
		LogPrefix:   consts.ProfileLogPrefixCreate,
	},
	consts.ProfileModeShow: {
		HeaderTitle: consts.ProfileUIHeaderShow,
		Mode:        consts.ProfileModeShow,
		SuccessMsg:  "", // No success message for show
		LogPrefix:   consts.ProfileLogPrefixShow,
	},
	consts.ProfileModeEdit: {
		HeaderTitle: consts.ProfileUIHeaderEdit,
		Mode:        consts.ProfileModeEdit,
		SuccessMsg:  consts.ProfileSuccessUpdated,
		LogPrefix:   consts.ProfileLogPrefixEdit,
	},
	consts.ProfileModeDelete: {
		HeaderTitle: consts.ProfileUIHeaderDelete,
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
	command, err := NewProfileCommand(config.Mode, cmd, deps, config)
	if err != nil {
		return err
	}
	return command.Execute()
}

func executeProfileCommon(cmd *cobra.Command, deps *appdeps.Dependencies, config profilemodel.ProfileEntryConfig, parse func() (interface{}, error)) error {
	logger := deps.Logger
	if config.LogPrefix != "" {
		logger.Infof(consts.ProfileLogStartProcessWithPrefixFmt, config.LogPrefix, config.Mode)
	} else {
		logger.Infof(consts.ProfileLogStartProcessFmt, config.Mode)
	}

	profileOptions, err := parse()
	if err != nil {
		return err
	}

	svc := NewProfileService(deps.Config, logger, profileOptions)

	if config.HeaderTitle != "" {
		print.PrintAppHeader(config.HeaderTitle)
	}

	if err := svc.ExecuteProfileCommand(config); err != nil {
		return err
	}

	if config.SuccessMsg != "" {
		print.PrintSuccess(config.SuccessMsg)
		logger.Info(config.SuccessMsg)
	}

	return nil
}

type createProfileCommand struct {
	cmd    *cobra.Command
	deps   *appdeps.Dependencies
	config profilemodel.ProfileEntryConfig
}

func (c *createProfileCommand) Execute() error {
	return executeProfileCommon(c.cmd, c.deps, c.config, func() (interface{}, error) {
		return parsing.ParsingCreateProfile(c.cmd, c.deps.Logger)
	})
}

type showProfileCommand struct {
	cmd    *cobra.Command
	deps   *appdeps.Dependencies
	config profilemodel.ProfileEntryConfig
}

func (c *showProfileCommand) Execute() error {
	return executeProfileCommon(c.cmd, c.deps, c.config, func() (interface{}, error) {
		return parsing.ParsingShowProfile(c.cmd)
	})
}

type editProfileCommand struct {
	cmd    *cobra.Command
	deps   *appdeps.Dependencies
	config profilemodel.ProfileEntryConfig
}

func (c *editProfileCommand) Execute() error {
	return executeProfileCommon(c.cmd, c.deps, c.config, func() (interface{}, error) {
		return parsing.ParsingEditProfile(c.cmd)
	})
}

type deleteProfileCommand struct {
	cmd    *cobra.Command
	deps   *appdeps.Dependencies
	config profilemodel.ProfileEntryConfig
}

func (c *deleteProfileCommand) Execute() error {
	return executeProfileCommon(c.cmd, c.deps, c.config, func() (interface{}, error) {
		return parsing.ParsingDeleteProfile(c.cmd)
	})
}

// NewProfileCommand membuat command executor untuk mode tertentu.
func NewProfileCommand(mode string, cmd *cobra.Command, deps *appdeps.Dependencies, config profilemodel.ProfileEntryConfig) (ProfileCommand, error) {
	switch mode {
	case consts.ProfileModeCreate:
		return &createProfileCommand{cmd: cmd, deps: deps, config: config}, nil
	case consts.ProfileModeShow:
		return &showProfileCommand{cmd: cmd, deps: deps, config: config}, nil
	case consts.ProfileModeEdit:
		return &editProfileCommand{cmd: cmd, deps: deps, config: config}, nil
	case consts.ProfileModeDelete:
		return &deleteProfileCommand{cmd: cmd, deps: deps, config: config}, nil
	default:
		return nil, profileerrors.ErrInvalidProfileMode
	}
}

// GetExecutionConfig mengembalikan konfigurasi untuk mode profile tertentu
func GetExecutionConfig(mode string) (profilemodel.ProfileEntryConfig, error) {
	config, ok := profileExecutionConfigs[mode]
	if !ok {
		return profilemodel.ProfileEntryConfig{}, profileerrors.ErrInvalidProfileMode
	}

	return config, nil
}
