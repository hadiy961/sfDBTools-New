// File : internal/restore/modes/single.go
// Deskripsi : Executor untuk restore single database
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package modes

import (
	"context"
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/ui"
	"time"
)

// SingleExecutor implements restore for single database
type SingleExecutor struct {
	service RestoreService
}

// NewSingleExecutor creates a new SingleExecutor
func NewSingleExecutor(svc RestoreService) *SingleExecutor {
	return &SingleExecutor{service: svc}
}

// Execute executes single database restore
func (e *SingleExecutor) Execute(ctx context.Context) (*types.RestoreResult, error) {
	startTime := time.Now()
	opts := e.service.GetSingleOptions()
	
	result := &types.RestoreResult{
		TargetDB:   opts.TargetDB,
		SourceFile: opts.File,
	}

	e.service.LogInfo("Memulai proses restore database")
	e.service.SetRestoreInProgress(opts.TargetDB)
	defer e.service.ClearRestoreInProgress()

	// 1. Check if database exists
	e.service.LogDebugf("Mengecek apakah database %s sudah ada...", opts.TargetDB)
	dbExists, err := e.service.GetTargetClient().CheckDatabaseExists(ctx, opts.TargetDB)
	if err != nil {
		result.Error = fmt.Errorf("gagal mengecek database target: %w", err)
		return result, result.Error
	}
	e.service.LogDebugf("Database %s exists: %v", opts.TargetDB, dbExists)

	// 2. Backup database if needed
	backupFile, err := e.service.BackupDatabaseIfNeeded(ctx, opts.TargetDB, dbExists, opts.SkipBackup, opts.BackupOptions)
	if err != nil {
		result.Error = err
		return result, result.Error
	}
	result.BackupFile = backupFile

	// 3. Drop database if needed
	if err := e.service.DropDatabaseIfNeeded(ctx, opts.TargetDB, dbExists, opts.DropTarget); err != nil {
		result.Error = err
		return result, result.Error
	}
	if opts.DropTarget && dbExists {
		result.DroppedDB = true
	}

	// 4. Create database and restore from file
	if err := e.service.CreateAndRestoreDatabase(ctx, opts.TargetDB, opts.File, opts.EncryptionKey); err != nil {
		result.Error = err
		return result, result.Error
	}

	// 5. Restore user grants if available
	result.GrantsFile = opts.GrantsFile
	grantsRestored, err := e.service.RestoreUserGrantsIfAvailable(ctx, opts.GrantsFile)
	if err != nil {
		e.service.LogErrorf("Gagal restore user grants: %v", err)
		ui.PrintWarning(fmt.Sprintf("⚠️  Database berhasil di-restore, tapi gagal restore user grants: %v", err))
		result.GrantsRestored = false
	} else {
		result.GrantsRestored = grantsRestored
	}

	result.Success = true
	result.Duration = time.Since(startTime).Round(time.Second).String()
	e.service.LogInfo("Restore database berhasil")

	return result, nil
}
