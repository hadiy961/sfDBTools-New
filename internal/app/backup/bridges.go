// File : internal/app/backup/bridges.go
// Deskripsi : Bridge methods untuk menghubungkan service backup dengan sub-modul
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2026-01-05

package backup

import (
	"context"
	"fmt"

	"sfdbtools/internal/app/backup/execution"
	"sfdbtools/internal/app/backup/grants"
	"sfdbtools/internal/app/backup/gtid"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/app/backup/modes"
	"sfdbtools/internal/app/backup/selection"
	"sfdbtools/internal/app/backup/setup"
	"sfdbtools/internal/shared/database"
)

// =============================================================================
// Selection bridges
// =============================================================================

func (s *Service) handleSingleModeSetup(ctx context.Context, client interface {
	GetDatabaseList(context.Context) ([]string, error)
}, dbFiltered []string) ([]string, error) {
	return selection.New(s.Log, s.BackupDBOptions).HandleSingleModeSetup(ctx, client, dbFiltered)
}

// =============================================================================
// Setup bridges
// =============================================================================

func (s *Service) SetupBackupExecution(state *BackupExecutionState) error {
	return setup.New(s.Log, s.Config, s.BackupDBOptions, &state.ExcludedDatabases).SetupBackupExecution()
}

func (s *Service) PrepareBackupSession(ctx context.Context, state *BackupExecutionState, headerTitle string, nonInteractive bool) (client *database.Client, dbFiltered []string, err error) {
	// Create closure that captures state for GenerateBackupPaths callback
	pathGenerator := func(ctx context.Context, client *database.Client, dbFiltered []string) ([]string, error) {
		return s.GenerateBackupPaths(ctx, state, client, dbFiltered)
	}
	return setup.New(s.Log, s.Config, s.BackupDBOptions, &state.ExcludedDatabases).
		PrepareBackupSession(ctx, headerTitle, nonInteractive, pathGenerator)
}

// =============================================================================
// Writer bridge
// =============================================================================

// =============================================================================
// Execution bridges
// =============================================================================

func (s *Service) ExecuteAndBuildBackup(ctx context.Context, state modes.BackupStateAccessor, cfg types_backup.BackupExecutionConfig) (types_backup.DatabaseBackupInfo, error) {
	// Type assert to access fields
	execState, ok := state.(*BackupExecutionState)
	if !ok {
		return types_backup.DatabaseBackupInfo{}, fmt.Errorf("invalid state type")
	}
	eng := execution.New(s.Log, s.Config, s.BackupDBOptions, s.ErrorLog).
		WithDependencies(s.Client, execState.GTIDInfo, execState.ExcludedDatabases, execState, s)
	return eng.ExecuteAndBuildBackup(ctx, cfg)
}

func (s *Service) ExecuteBackupLoop(ctx context.Context, state modes.BackupStateAccessor, databases []string, config types_backup.BackupLoopConfig, outputPathFunc func(dbName string) (string, error)) types_backup.BackupLoopResult {
	// Type assert to access fields
	execState, ok := state.(*BackupExecutionState)
	if !ok {
		return types_backup.BackupLoopResult{
			Errors: []string{"invalid state type"},
		}
	}
	eng := execution.New(s.Log, s.Config, s.BackupDBOptions, s.ErrorLog).
		WithDependencies(s.Client, execState.GTIDInfo, execState.ExcludedDatabases, execState, s)
	return eng.ExecuteBackupLoop(ctx, databases, config, outputPathFunc)
}

// =============================================================================
// Grants bridges
// =============================================================================

func (s *Service) ExportUserGrantsIfNeeded(ctx context.Context, referenceBackupFile string, databases []string) string {
	return grants.ExportUserGrantsIfNeeded(ctx, s.Client, s.Log, referenceBackupFile, s.BackupDBOptions.ExcludeUser, s.BackupDBOptions.DryRun, databases)
}

func (s *Service) UpdateMetadataUserGrantsPath(backupFilePath string, userGrantsPath string) {
	grants.UpdateMetadataUserGrantsPath(s.Log, backupFilePath, userGrantsPath)
}

// =============================================================================
// GTID bridge
// =============================================================================

func (s *Service) CaptureAndSaveGTID(ctx context.Context, state modes.BackupStateAccessor, backupFilePath string) error {
	gtidInfo, err := gtid.Capture(ctx, s.Client, s.Log, s.BackupDBOptions.CaptureGTID)
	if err != nil {
		return err
	}
	if gtidInfo != nil {
		// Type assert state to *BackupExecutionState to access GTIDInfo field
		if execState, ok := state.(*BackupExecutionState); ok {
			execState.GTIDInfo = gtidInfo
		}
	}
	return nil
}
