// File : internal/restore/modes/single.go
// Deskripsi : Executor untuk restore single database
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package modes

import (
	"context"
	"fmt"
	"sfDBTools/internal/restore/helpers"
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

	logger := e.service.GetLogger()
	logger.Info("Memulai proses restore database")
	e.service.SetRestoreInProgress(opts.TargetDB)
	defer e.service.ClearRestoreInProgress()

	// Dry-run mode: validasi file tanpa restore
	if opts.DryRun {
		logger.Info("Mode DRY-RUN: Validasi file tanpa restore...")
		return e.executeDryRun(ctx, opts, result, startTime)
	}

	// 1. Check if database exists
	logger.Debugf("Mengecek apakah database %s sudah ada...", opts.TargetDB)
	dbExists, err := e.service.GetTargetClient().CheckDatabaseExists(ctx, opts.TargetDB)
	if err != nil {
		result.Error = fmt.Errorf("gagal mengecek database target: %w", err)
		return result, result.Error
	}
	logger.Debugf("Database %s exists: %v", opts.TargetDB, dbExists)

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
		logger.Errorf("Gagal restore user grants: %v", err)
		ui.PrintWarning(fmt.Sprintf("⚠️  Database berhasil di-restore, tapi gagal restore user grants: %v", err))
		result.GrantsRestored = false
	} else {
		result.GrantsRestored = grantsRestored
	}

	result.Success = true
	result.Duration = time.Since(startTime).Round(time.Second).String()
	logger.Info("Restore database berhasil")

	return result, nil
}

// executeDryRun melakukan validasi file backup tanpa restore
func (e *SingleExecutor) executeDryRun(ctx context.Context, opts *types.RestoreSingleOptions, result *types.RestoreResult, startTime time.Time) (*types.RestoreResult, error) {
	logger := e.service.GetLogger()
	logger.Info("Validasi file backup...")

	// Validasi file exist dan bisa dibaca
	reader, closers, err := helpers.OpenAndPrepareReader(opts.File, opts.EncryptionKey)
	if err != nil {
		result.Error = fmt.Errorf("gagal membuka file: %w", err)
		return result, result.Error
	}
	defer helpers.CloseReaders(closers)

	// Close reader immediately setelah validasi
	_ = reader

	// Check database status
	dbExists, err := e.service.GetTargetClient().CheckDatabaseExists(ctx, opts.TargetDB)
	if err != nil {
		result.Error = fmt.Errorf("gagal mengecek database target: %w", err)
		return result, result.Error
	}

	// Print hasil validasi
	ui.PrintSuccess("\n✓ Validasi File Backup:")
	ui.PrintInfo(fmt.Sprintf("  Source File: %s", opts.File))
	ui.PrintInfo(fmt.Sprintf("  Target DB: %s", opts.TargetDB))
	ui.PrintInfo(fmt.Sprintf("  DB Exists: %v", dbExists))
	if dbExists && !opts.SkipBackup {
		ui.PrintInfo("  Pre-restore Backup: Will be created")
	}
	if opts.DropTarget && dbExists {
		ui.PrintWarning("  ⚠️  Database will be DROPPED before restore")
	}

	result.Success = true
	result.Duration = time.Since(startTime).Round(time.Second).String()
	return result, nil
}
