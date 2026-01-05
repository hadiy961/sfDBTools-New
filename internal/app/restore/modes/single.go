// File : internal/restore/modes/single.go
// Deskripsi : Executor untuk restore single database
// Author : Hadiyatna Muflihun
// Tanggal : 17 Desember 2025
// Last Modified : 5 Januari 2026
package modes

import (
	"context"
	"fmt"
	restoremodel "sfDBTools/internal/app/restore/model"
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
func (e *SingleExecutor) Execute(ctx context.Context) (*restoremodel.RestoreResult, error) {
	startTime := time.Now()
	opts := e.service.GetSingleOptions()
	logger := e.service.GetLogger()

	result := createResultWithDefaults(opts.TargetDB, opts.File, startTime)

	logger.Info("Memulai proses restore database")
	e.service.SetRestoreInProgress(opts.TargetDB)
	defer e.service.ClearRestoreInProgress()

	// Dry-run mode: validasi file tanpa restore
	if opts.DryRun {
		logger.Info("Mode DRY-RUN: Validasi file tanpa restore...")
		return e.executeDryRun(ctx, opts, result, startTime)
	}

	// Execute common restore flow (backup -> drop -> restore)
	flow := &commonRestoreFlow{
		service:       e.service,
		ctx:           ctx,
		dbName:        opts.TargetDB,
		sourceFile:    opts.File,
		encryptionKey: opts.EncryptionKey,
		skipBackup:    opts.SkipBackup,
		dropTarget:    opts.DropTarget,
		backupOpts:    opts.BackupOptions,
	}

	backupFile, err := flow.execute()
	if err != nil {
		result.Error = err
		return result, err
	}
	result.BackupFile = backupFile

	// Restore user grants if available
	result.GrantsFile = opts.GrantsFile
	result.GrantsRestored = performGrantsRestore(ctx, e.service, opts.GrantsFile, false)

	// Post-restore operations (temp DB + grants copy)
	performPostRestoreOperations(ctx, e.service, opts.TargetDB)

	finalizeResult(result, startTime, true)
	logger.Info("Restore database berhasil")

	return result, nil
}

// executeDryRun melakukan validasi file backup tanpa restore
func (e *SingleExecutor) executeDryRun(ctx context.Context, opts *restoremodel.RestoreSingleOptions, result *restoremodel.RestoreResult, startTime time.Time) (*restoremodel.RestoreResult, error) {
	validator := newDryRunValidator(e.service, ctx, result, startTime)

	// Validate file can be opened
	if err := validator.validateSingleFile(opts.File, opts.EncryptionKey); err != nil {
		result.Error = err
		return result, err
	}

	// Check database status
	dbExists, err := validator.validateDatabaseStatus(opts.TargetDB)
	if err != nil {
		result.Error = err
		return result, err
	}

	// Build summary info
	info := map[string]string{
		"Source File": opts.File,
		"Target DB":   opts.TargetDB,
		"DB Exists":   fmt.Sprintf("%v", dbExists),
	}
	if dbExists && !opts.SkipBackup {
		info["Pre-restore Backup"] = "Will be created"
	}

	warnings := []string{}
	if opts.DropTarget && dbExists {
		warnings = append(warnings, "⚠️  Database will be DROPPED before restore")
	}

	validator.printSummary(info, warnings)
	return validator.finalize()
}
