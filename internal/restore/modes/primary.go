// File : internal/restore/modes/primary.go
// Deskripsi : Executor untuk restore primary database dengan companion
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

// PrimaryExecutor implements restore for primary database with companion
type PrimaryExecutor struct {
	service RestoreService
}

// NewPrimaryExecutor creates a new PrimaryExecutor
func NewPrimaryExecutor(svc RestoreService) *PrimaryExecutor {
	return &PrimaryExecutor{service: svc}
}

// Execute executes primary database restore with companion
func (e *PrimaryExecutor) Execute(ctx context.Context) (*types.RestoreResult, error) {
	startTime := time.Now()
	opts := e.service.GetPrimaryOptions()

	result := &types.RestoreResult{
		TargetDB:   opts.TargetDB,
		SourceFile: opts.File,
	}

	e.service.LogInfo("Memulai proses restore database primary")
	e.service.SetRestoreInProgress(opts.TargetDB)
	defer e.service.ClearRestoreInProgress()

	// 1. Check if database exists
	dbExists, err := e.service.GetTargetClient().CheckDatabaseExists(ctx, opts.TargetDB)
	if err != nil {
		result.Error = fmt.Errorf("gagal mengecek database target: %w", err)
		return result, result.Error
	}

	// 2. Detect companion file if needed
	if opts.IncludeDmart {
		if err := e.service.DetectOrSelectCompanionFile(); err != nil {
			result.Error = fmt.Errorf("gagal deteksi companion database: %w", err)
			return result, result.Error
		}
		// Refresh opts after potential update in DetectOrSelectCompanionFile (pointer reflected)
		// Note: DetectOrSelectCompanionFile modifies the struct inside service.
		result.CompanionFile = opts.CompanionFile
		result.CompanionDB = opts.TargetDB + "_dmart"
	}

	// 3. Backup primary database if needed
	backupFile, err := e.service.BackupDatabaseIfNeeded(ctx, opts.TargetDB, dbExists, opts.SkipBackup, opts.BackupOptions)
	if err != nil {
		result.Error = err
		return result, result.Error
	}
	result.BackupFile = backupFile

	// Backup companion if needed
	if opts.IncludeDmart && !opts.SkipBackup {
		companionDB := opts.TargetDB + "_dmart"
		companionExists, err := e.service.GetTargetClient().CheckDatabaseExists(ctx, companionDB)
		if err == nil && companionExists {
			companionBackup, err := e.service.BackupDatabaseIfNeeded(ctx, companionDB, true, false, opts.BackupOptions)
			if err != nil {
				e.service.LogWarnf("Gagal backup companion database: %v", err)
			} else if companionBackup != "" {
				result.CompanionBackup = companionBackup
			}
		}
	}

	// 4. Drop primary database if needed
	if err := e.service.DropDatabaseIfNeeded(ctx, opts.TargetDB, dbExists, opts.DropTarget); err != nil {
		result.Error = err
		return result, result.Error
	}
	if opts.DropTarget && dbExists {
		result.DroppedDB = true
	}

	// Drop companion if needed
	if opts.IncludeDmart && opts.DropTarget {
		companionDB := opts.TargetDB + "_dmart"
		companionExists, err := e.service.GetTargetClient().CheckDatabaseExists(ctx, companionDB)
		if err == nil && companionExists {
			if err := e.service.DropDatabaseIfNeeded(ctx, companionDB, true, true); err != nil {
				e.service.LogWarnf("Gagal drop companion database: %v", err)
			} else {
				result.DroppedCompanion = true
			}
		}
	}

	// 5. Create and restore primary database
	if err := e.service.CreateAndRestoreDatabase(ctx, opts.TargetDB, opts.File, opts.EncryptionKey); err != nil {
		result.Error = err
		return result, result.Error
	}

	// 6. Restore companion database if available
	if opts.IncludeDmart && opts.CompanionFile != "" {
		companionDB := opts.TargetDB + "_dmart"
		e.service.LogInfof("Restore companion database dari %s...", opts.CompanionFile)

		if err := e.service.CreateAndRestoreDatabase(ctx, companionDB, opts.CompanionFile, opts.EncryptionKey); err != nil {
			e.service.LogWarnf("Gagal restore companion database: %v", err)
			ui.PrintWarning(fmt.Sprintf("⚠️  Gagal restore companion database %s: %v", companionDB, err))
		} else {
			e.service.LogInfof("Companion database %s berhasil di-restore", companionDB)
		}
	}

	// 7. Restore user grants if available
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
	e.service.LogInfo("Restore primary database berhasil")

	return result, nil
}
