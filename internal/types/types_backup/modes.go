// File : internal/types/types_backup/modes.go
// Deskripsi : Mode interface dan config structs
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package types_backup

import (
	"context"
	"sfDBTools/internal/types"
)

// ModeExecutor interface untuk semua mode backup
type ModeExecutor interface {
	Execute(ctx context.Context, databases []string) BackupResult
}

// BackupService interface untuk service yang dibutuhkan oleh mode executors
// Ini memisahkan concerns dan membuat mode executors tidak tightly coupled ke Service
type BackupService interface {
	// Logging methods
	LogInfo(msg string)
	LogDebug(msg string)
	LogWarn(msg string)
	LogError(msg string)

	// Backup execution
	ExecuteAndBuildBackup(ctx context.Context, cfg BackupExecutionConfig) (types.DatabaseBackupInfo, error)
	ExecuteBackupLoop(ctx context.Context, databases []string, config BackupLoopConfig, outputPathFunc func(dbName string) (string, error)) BackupLoopResult

	// Helper methods
	GetBackupOptions() *BackupDBOptions
	GenerateFullBackupPath(dbName string, mode string) (string, error)
	GetTotalDatabaseCount(ctx context.Context, dbFiltered []string) int
	CaptureAndSaveGTID(ctx context.Context, backupFilePath string) error
	ExportUserGrantsIfNeeded(ctx context.Context, referenceBackupFile string)
	ToBackupResult(loopResult BackupLoopResult) BackupResult
}

// BackupLoopConfig konfigurasi untuk backup loop execution
type BackupLoopConfig struct {
	Mode       string // "single" atau "separated"
	TotalDBs   int    // Total database untuk progress display
	BackupType string // "single" atau "separated"
}

// BackupLoopResult hasil dari loop execution
type BackupLoopResult struct {
	Success     int
	Failed      int
	BackupInfos []types.DatabaseBackupInfo
	FailedDBs   []FailedDatabaseInfo
	Errors      []string
}
