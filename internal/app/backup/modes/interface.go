// File : internal/backup/modes/interface.go
// Deskripsi : Interface definitions untuk backup modes (ISP-compliant)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 20 Januari 2026
package modes

import (
	"context"
	"sfdbtools/internal/app/backup/model/types_backup"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
)

// ModeExecutor interface untuk semua mode backup
type ModeExecutor interface {
	Execute(ctx context.Context, databases []string) types_backup.BackupResult
}

// BackupStateAccessor menyediakan akses ke execution state untuk mode executors.
// Ini memungkinkan mode executors untuk berinteraksi dengan state tanpa coupling ke Service.
type BackupStateAccessor interface {
	// SetCurrentBackupFile mencatat file backup yang sedang dibuat
	SetCurrentBackupFile(filePath string)
	// ClearCurrentBackupFile menghapus catatan file backup setelah selesai
	ClearCurrentBackupFile()
	// GetCurrentBackupFile returns current backup file path dan status
	GetCurrentBackupFile() (string, bool)
}

// =============================================================================
// ISP-Compliant Interfaces (Interface Segregation Principle)
// Setiap interface memiliki 1-3 methods untuk cohesion yang tinggi
// =============================================================================

// BackupContext menyediakan akses ke logger dan configuration
type BackupContext interface {
	GetLog() applog.Logger
	GetConfig() *appconfig.Config
	GetOptions() *types_backup.BackupDBOptions
}

// BackupExecutor menangani core backup execution logic
type BackupExecutor interface {
	ExecuteAndBuildBackup(ctx context.Context, state BackupStateAccessor, cfg types_backup.BackupExecutionConfig) (types_backup.DatabaseBackupInfo, error)
	ExecuteBackupLoop(ctx context.Context, state BackupStateAccessor, databases []string, config types_backup.BackupLoopConfig, outputPathFunc func(dbName string) (string, error)) types_backup.BackupLoopResult
}

// BackupPathProvider menghasilkan path untuk backup files
type BackupPathProvider interface {
	GenerateFullBackupPath(dbName string, mode string) (string, error)
}

// BackupMetadata menangani metadata operations (GTID, user grants)
type BackupMetadata interface {
	CaptureAndSaveGTID(ctx context.Context, state BackupStateAccessor, backupFilePath string) error
	ExportUserGrantsIfNeeded(ctx context.Context, referenceBackupFile string, databases []string) string
	UpdateMetadataUserGrantsPath(backupFilePath string, userGrantsPath string, permissions string)
}

// BackupResultConverter mengkonversi loop result ke backup result
type BackupResultConverter interface {
	ToBackupResult(loopResult types_backup.BackupLoopResult) types_backup.BackupResult
}

// BackupService adalah composite interface yang menggabungkan semua capabilities
// Mode executors dapat depend pada interface ini atau sub-interface yang lebih spesifik
type BackupService interface {
	BackupContext
	BackupExecutor
	BackupPathProvider
	BackupMetadata
	BackupResultConverter
}
