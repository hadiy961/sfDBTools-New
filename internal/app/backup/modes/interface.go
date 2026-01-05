// File : internal/backup/modes/interface.go
// Deskripsi : Interface definitions untuk backup modes (ISP-compliant)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2026-01-05
package modes

import (
	"context"
	"sfDBTools/internal/app/backup/model/types_backup"
	applog "sfDBTools/internal/services/log"
)

// ModeExecutor interface untuk semua mode backup
type ModeExecutor interface {
	Execute(ctx context.Context, databases []string) types_backup.BackupResult
}

// =============================================================================
// ISP-Compliant Interfaces (Interface Segregation Principle)
// Setiap interface memiliki 1-3 methods untuk cohesion yang tinggi
// =============================================================================

// BackupContext menyediakan akses ke logger dan configuration
type BackupContext interface {
	GetLog() applog.Logger
	GetOptions() *types_backup.BackupDBOptions
}

// BackupExecutor menangani core backup execution logic
type BackupExecutor interface {
	ExecuteAndBuildBackup(ctx context.Context, cfg types_backup.BackupExecutionConfig) (types_backup.DatabaseBackupInfo, error)
	ExecuteBackupLoop(ctx context.Context, databases []string, config types_backup.BackupLoopConfig, outputPathFunc func(dbName string) (string, error)) types_backup.BackupLoopResult
}

// BackupPathProvider menghasilkan path untuk backup files
type BackupPathProvider interface {
	GenerateFullBackupPath(dbName string, mode string) (string, error)
}

// BackupMetadata menangani metadata operations (GTID, user grants)
type BackupMetadata interface {
	CaptureAndSaveGTID(ctx context.Context, backupFilePath string) error
	ExportUserGrantsIfNeeded(ctx context.Context, referenceBackupFile string, databases []string) string
	UpdateMetadataUserGrantsPath(backupFilePath string, userGrantsPath string)
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
