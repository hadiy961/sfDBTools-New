// File : internal/restore/modes/interface.go
// Deskripsi : Interface dan type definitions untuk restore modes
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified :  2026-01-05
package modes

import (
	"context"
	restoremodel "sfDBTools/internal/app/restore/model"
	"sfDBTools/internal/domain"
	"sfDBTools/internal/services/log"
	"sfDBTools/pkg/database"
)

// RestoreExecutor interface untuk semua mode restore
type RestoreExecutor interface {
	Execute(ctx context.Context) (*restoremodel.RestoreResult, error)
}

// RestoreService interface untuk service yang dibutuhkan oleh mode executors
type RestoreService interface {
	// Logging
	GetLogger() applog.Logger

	// Context & Clients
	GetTargetClient() *database.Client
	GetProfile() *domain.ProfileInfo

	// State Management
	SetRestoreInProgress(dbName string)
	ClearRestoreInProgress()

	// Options Accessors
	GetSingleOptions() *restoremodel.RestoreSingleOptions
	GetPrimaryOptions() *restoremodel.RestorePrimaryOptions
	GetSecondaryOptions() *restoremodel.RestoreSecondaryOptions
	GetAllOptions() *restoremodel.RestoreAllOptions
	GetSelectionOptions() *restoremodel.RestoreSelectionOptions
	GetCustomOptions() *restoremodel.RestoreCustomOptions

	// Restore Operations (Exposed from helpers)
	BackupDatabaseIfNeeded(ctx context.Context, dbName string, dbExists bool, skipBackup bool, backupOpts *restoremodel.RestoreBackupOptions) (string, error)
	// BackupDatabasesSingleFileIfNeeded melakukan backup gabungan (single-file/combined)
	// untuk sekumpulan database sebelum restore all (konsep: backup filter --mode single-file).
	BackupDatabasesSingleFileIfNeeded(ctx context.Context, dbNames []string, skipBackup bool, backupOpts *restoremodel.RestoreBackupOptions) (string, error)
	DropDatabaseIfNeeded(ctx context.Context, dbName string, dbExists bool, shouldDrop bool) error
	CreateAndRestoreDatabase(ctx context.Context, dbName string, filePath string, encryptionKey string) error
	RestoreUserGrantsIfAvailable(ctx context.Context, grantsFile string) (bool, error)
	// Post-restore operations (best-effort; caller may treat errors as warnings)
	CreateTempDatabaseIfNeeded(ctx context.Context, dbName string) (string, error)
	CopyDatabaseGrants(ctx context.Context, sourceDB string, targetDB string) error
	DetectOrSelectCompanionFile() error
}
