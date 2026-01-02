// File : internal/backup/bridges.go
// Deskripsi : Bridge methods untuk menghubungkan service backup dengan sub-modul
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2026-01-02

package backup

import (
	"context"

	"sfDBTools/internal/backup/execution"
	"sfDBTools/internal/backup/grants"
	"sfDBTools/internal/backup/gtid"
	"sfDBTools/internal/backup/selection"
	"sfDBTools/internal/backup/setup"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/database"
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

func (s *Service) CheckAndSelectConfigFile() error {
	return setup.New(s.Log, s.Config, s.BackupDBOptions, &s.excludedDatabases).CheckAndSelectConfigFile()
}

func (s *Service) SetupBackupExecution() error {
	return setup.New(s.Log, s.Config, s.BackupDBOptions, &s.excludedDatabases).SetupBackupExecution()
}

func (s *Service) GetFilteredDatabases(ctx context.Context, client *database.Client) ([]string, *types.FilterStats, error) {
	return setup.New(s.Log, s.Config, s.BackupDBOptions, &s.excludedDatabases).GetFilteredDatabases(ctx, client)
}

func (s *Service) PrepareBackupSession(ctx context.Context, headerTitle string, nonInteractive bool) (client *database.Client, dbFiltered []string, err error) {
	return setup.New(s.Log, s.Config, s.BackupDBOptions, &s.excludedDatabases).
		PrepareBackupSession(ctx, headerTitle, nonInteractive, s.GenerateBackupPaths)
}

// =============================================================================
// Writer bridge
// =============================================================================

// =============================================================================
// Execution bridges
// =============================================================================

func (s *Service) ExecuteAndBuildBackup(ctx context.Context, cfg types_backup.BackupExecutionConfig) (types_backup.DatabaseBackupInfo, error) {
	eng := execution.New(s.Log, s.Config, s.BackupDBOptions, s.ErrorLog).
		WithDependencies(s.Client, s.gtidInfo, s.excludedDatabases, s, s)
	return eng.ExecuteAndBuildBackup(ctx, cfg)
}

func (s *Service) ExecuteBackupLoop(ctx context.Context, databases []string, config types_backup.BackupLoopConfig, outputPathFunc func(dbName string) (string, error)) types_backup.BackupLoopResult {
	eng := execution.New(s.Log, s.Config, s.BackupDBOptions, s.ErrorLog).
		WithDependencies(s.Client, s.gtidInfo, s.excludedDatabases, s, s)
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

func (s *Service) CaptureAndSaveGTID(ctx context.Context, backupFilePath string) error {
	gtidInfo, err := gtid.Capture(ctx, s.Client, s.Log, s.BackupDBOptions.CaptureGTID)
	if err != nil {
		return err
	}
	if gtidInfo != nil {
		s.gtidInfo = gtidInfo
	}
	return nil
}
