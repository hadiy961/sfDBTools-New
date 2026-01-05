// File : internal/restore/modes/interface.go
// Deskripsi : Interface dan type definitions untuk restore modes
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified :  2026-01-05
package modes

import (
	"context"
	"sfDBTools/internal/services/log"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
)

// RestoreExecutor interface untuk semua mode restore
type RestoreExecutor interface {
	Execute(ctx context.Context) (*types.RestoreResult, error)
}

// RestoreService interface untuk service yang dibutuhkan oleh mode executors
type RestoreService interface {
	// Logging
	GetLogger() applog.Logger

	// Context & Clients
	GetTargetClient() *database.Client
	GetProfile() *types.ProfileInfo

	// State Management
	SetRestoreInProgress(dbName string)
	ClearRestoreInProgress()

	// Options Accessors
	GetSingleOptions() *types.RestoreSingleOptions
	GetPrimaryOptions() *types.RestorePrimaryOptions
	GetSecondaryOptions() *types.RestoreSecondaryOptions
	GetAllOptions() *types.RestoreAllOptions
	GetSelectionOptions() *types.RestoreSelectionOptions
	GetCustomOptions() *types.RestoreCustomOptions

	// Restore Operations (Exposed from helpers)
	BackupDatabaseIfNeeded(ctx context.Context, dbName string, dbExists bool, skipBackup bool, backupOpts *types.RestoreBackupOptions) (string, error)
	// BackupDatabasesSingleFileIfNeeded melakukan backup gabungan (single-file/combined)
	// untuk sekumpulan database sebelum restore all (konsep: backup filter --mode single-file).
	BackupDatabasesSingleFileIfNeeded(ctx context.Context, dbNames []string, skipBackup bool, backupOpts *types.RestoreBackupOptions) (string, error)
	DropDatabaseIfNeeded(ctx context.Context, dbName string, dbExists bool, shouldDrop bool) error
	CreateAndRestoreDatabase(ctx context.Context, dbName string, filePath string, encryptionKey string) error
	RestoreUserGrantsIfAvailable(ctx context.Context, grantsFile string) (bool, error)
	// Post-restore operations (best-effort; caller may treat errors as warnings)
	CreateTempDatabaseIfNeeded(ctx context.Context, dbName string) (string, error)
	CopyDatabaseGrants(ctx context.Context, sourceDB string, targetDB string) error
	DetectOrSelectCompanionFile() error
}
