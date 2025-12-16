// File : internal/backup/modes/interface.go
// Deskripsi : Interface dan type definitions untuk backup modes
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package modes

import (
	"context"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
)

// ModeExecutor interface untuk semua mode backup
type ModeExecutor interface {
	Execute(ctx context.Context, databases []string) types_backup.BackupResult
}

// BackupService interface untuk service yang dibutuhkan oleh mode executors
// Ini memisahkan concerns dan membuat mode executors tidak tightly coupled ke Service
type BackupService interface {
	// Logging methods
	LogInfo(msg string)
	LogDebug(msg string)
	LogWarn(msg string)
	LogError(msg string)
	GetLogger() applog.Logger

	// Backup execution
	ExecuteAndBuildBackup(ctx context.Context, cfg types_backup.BackupExecutionConfig) (types.DatabaseBackupInfo, error)
	ExecuteBackupLoop(ctx context.Context, databases []string, config types_backup.BackupLoopConfig, outputPathFunc func(dbName string) (string, error)) types_backup.BackupLoopResult

	// Helper methods
	GetBackupOptions() *types_backup.BackupDBOptions
	GenerateFullBackupPath(dbName string, mode string) (string, error)
	GetTotalDatabaseCount(ctx context.Context, dbFiltered []string) int
	CaptureAndSaveGTID(ctx context.Context, backupFilePath string) error
	ExportUserGrantsIfNeeded(ctx context.Context, referenceBackupFile string, databases []string) string
	UpdateMetadataUserGrantsPath(backupFilePath string, userGrantsPath string)
	ToBackupResult(loopResult types_backup.BackupLoopResult) types_backup.BackupResult
}
