package backup

import (
	"context"

	"sfDBTools/internal/backup/execution"
	"sfDBTools/internal/backup/grants"
	"sfDBTools/internal/backup/gtid"
	"sfDBTools/internal/backup/selection"
	"sfDBTools/internal/backup/setup"
	"sfDBTools/internal/backup/writer"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/database"
)

// =============================================================================
// Selection bridges
// =============================================================================

func (s *Service) getFilteredDatabasesWithMultiSelect(ctx context.Context, client *database.Client) ([]string, *types.FilterStats, error) {
	return selection.New(s.Log, s.BackupDBOptions).GetFilteredDatabasesWithMultiSelect(ctx, client)
}

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

func (s *Service) PrepareBackupSession(ctx context.Context, headerTitle string, showOptions bool) (client *database.Client, dbFiltered []string, err error) {
	return setup.New(s.Log, s.Config, s.BackupDBOptions, &s.excludedDatabases).
		PrepareBackupSession(ctx, headerTitle, showOptions, s.GenerateBackupPaths)
}

// =============================================================================
// Writer bridge
// =============================================================================

func (s *Service) executeMysqldumpWithPipe(ctx context.Context, mysqldumpArgs []string, outputPath string, compressionRequired bool, compressionType string) (*types_backup.BackupWriteResult, error) {
	return writer.New(s.Log, s.ErrorLog, s.BackupDBOptions).
		ExecuteMysqldumpWithPipe(ctx, mysqldumpArgs, outputPath, compressionRequired, compressionType)
}

// =============================================================================
// Execution bridges
// =============================================================================

func (s *Service) ExecuteAndBuildBackup(ctx context.Context, cfg types_backup.BackupExecutionConfig) (types_backup.DatabaseBackupInfo, error) {
	eng := execution.New(s.Log, s.Config, s.BackupDBOptions, s.ErrorLog)
	eng.Client = s.Client
	eng.GTIDInfo = s.gtidInfo
	eng.ExcludedDatabases = s.excludedDatabases
	eng.State = s
	eng.UserGrants = s
	return eng.ExecuteAndBuildBackup(ctx, cfg)
}

func (s *Service) ExecuteBackupLoop(ctx context.Context, databases []string, config types_backup.BackupLoopConfig, outputPathFunc func(dbName string) (string, error)) types_backup.BackupLoopResult {
	eng := execution.New(s.Log, s.Config, s.BackupDBOptions, s.ErrorLog)
	eng.Client = s.Client
	eng.GTIDInfo = s.gtidInfo
	eng.ExcludedDatabases = s.excludedDatabases
	eng.State = s
	eng.UserGrants = s
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
