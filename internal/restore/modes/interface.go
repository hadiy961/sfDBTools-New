// File : internal/restore/modes/interface.go
// Deskripsi : Interface dan type definitions untuk restore modes
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package modes

import (
	"context"
	"sfDBTools/internal/applog"
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
	GetAllOptions() *types.RestoreAllOptions
	GetSelectionOptions() *types.RestoreSelectionOptions

	// Restore Operations (Exposed from helpers)
	BackupDatabaseIfNeeded(ctx context.Context, dbName string, dbExists bool, skipBackup bool, backupOpts *types.RestoreBackupOptions) (string, error)
	DropDatabaseIfNeeded(ctx context.Context, dbName string, dbExists bool, shouldDrop bool) error
	CreateAndRestoreDatabase(ctx context.Context, dbName string, filePath string, encryptionKey string) error
	RestoreUserGrantsIfAvailable(ctx context.Context, grantsFile string) (bool, error)
	DetectOrSelectCompanionFile() error
}
