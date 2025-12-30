// File : internal/restore/command.go
// Deskripsi : Command execution functions untuk cmd layer
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-30

package restore

import (
	appdeps "sfDBTools/internal/deps"
	"sfDBTools/internal/parsing"
	"sfDBTools/internal/restore/display"

	"github.com/spf13/cobra"
)

// ExecuteRestoreCustomCommand adalah entry point untuk restore custom (SFCola account detail)
func ExecuteRestoreCustomCommand(cmd *cobra.Command, deps *appdeps.Dependencies) error {
	return executeRestoreCommand(
		cmd,
		deps,
		"Memulai proses restore custom",
		func(cmd *cobra.Command) (interface{}, error) {
			parsedOpts, err := parsing.ParsingRestoreCustomOptions(cmd)
			if err != nil {
				return nil, err
			}
			return &parsedOpts, nil
		},
		(*Service).SetupRestoreCustomSession,
		(*Service).ExecuteRestoreCustom,
		display.ShowRestoreCustomResult,
		"Proses restore custom dibatalkan.",
		"Restore custom gagal: ",
		"ExecuteRestoreCustom",
		"Restore custom berhasil diselesaikan",
	)
}

// ExecuteRestoreSingleCommand adalah entry point untuk restore single command
func ExecuteRestoreSingleCommand(cmd *cobra.Command, deps *appdeps.Dependencies) error {
	return executeRestoreCommand(
		cmd,
		deps,
		"Memulai proses restore single database",
		func(cmd *cobra.Command) (interface{}, error) {
			parsedOpts, err := parsing.ParsingRestoreSingleOptions(cmd)
			if err != nil {
				return nil, err
			}
			return &parsedOpts, nil
		},
		(*Service).SetupRestoreSession,
		(*Service).ExecuteRestoreSingle,
		display.ShowRestoreSingleResult,
		"Proses restore single dibatalkan.",
		"Restore single gagal: ",
		"ExecuteRestoreSingle",
		"Restore database berhasil diselesaikan",
	)
}

// ExecuteRestorePrimaryCommand adalah entry point untuk restore primary command
func ExecuteRestorePrimaryCommand(cmd *cobra.Command, deps *appdeps.Dependencies) error {
	return executeRestoreCommand(
		cmd,
		deps,
		"Memulai proses restore primary database",
		func(cmd *cobra.Command) (interface{}, error) {
			parsedOpts, err := parsing.ParsingRestorePrimaryOptions(cmd)
			if err != nil {
				return nil, err
			}
			return &parsedOpts, nil
		},
		(*Service).SetupRestorePrimarySession,
		(*Service).ExecuteRestorePrimary,
		display.ShowRestorePrimaryResult,
		"Proses restore primary dibatalkan.",
		"Restore primary gagal: ",
		"ExecuteRestorePrimary",
		"Restore primary database berhasil diselesaikan",
	)
}

// ExecuteRestoreSecondaryCommand adalah entry point untuk restore secondary command
func ExecuteRestoreSecondaryCommand(cmd *cobra.Command, deps *appdeps.Dependencies) error {
	return executeRestoreCommand(
		cmd,
		deps,
		"Memulai proses restore secondary database",
		func(cmd *cobra.Command) (interface{}, error) {
			parsedOpts, err := parsing.ParsingRestoreSecondaryOptions(cmd)
			if err != nil {
				return nil, err
			}
			return &parsedOpts, nil
		},
		(*Service).SetupRestoreSecondarySession,
		(*Service).ExecuteRestoreSecondary,
		display.ShowRestoreSecondaryResult,
		"Proses restore secondary dibatalkan.",
		"Restore secondary gagal: ",
		"ExecuteRestoreSecondary",
		"Restore secondary database berhasil diselesaikan",
	)
}

// ExecuteRestoreAllCommand adalah entry point untuk restore all databases command
func ExecuteRestoreAllCommand(cmd *cobra.Command, deps *appdeps.Dependencies) error {
	return executeRestoreCommand(
		cmd,
		deps,
		"Memulai proses restore all databases",
		func(cmd *cobra.Command) (interface{}, error) {
			parsedOpts, err := parsing.ParsingRestoreAllOptions(cmd)
			if err != nil {
				return nil, err
			}
			return &parsedOpts, nil
		},
		(*Service).SetupRestoreAllSession,
		(*Service).ExecuteRestoreAll,
		display.ShowRestoreAllResult,
		"Proses restore all dibatalkan.",
		"Restore all gagal: ",
		"ExecuteRestoreAll",
		"Restore all databases berhasil diselesaikan",
	)
}

// ExecuteRestoreSelectionCommand adalah entry point untuk restore selection (CSV)
func ExecuteRestoreSelectionCommand(cmd *cobra.Command, deps *appdeps.Dependencies) error {
	return executeRestoreCommand(
		cmd,
		deps,
		"Memulai proses restore selection (CSV)",
		func(cmd *cobra.Command) (interface{}, error) {
			parsedOpts, err := parsing.ParsingRestoreSelectionOptions(cmd)
			if err != nil {
				return nil, err
			}
			return &parsedOpts, nil
		},
		(*Service).SetupRestoreSelectionSession,
		(*Service).ExecuteRestoreSelection,
		display.ShowRestoreAllResult,
		"Proses restore selection dibatalkan.",
		"Restore selection gagal: ",
		"ExecuteRestoreSelection",
		"Restore selection selesai",
	)
}
