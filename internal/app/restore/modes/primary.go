// File : internal/restore/modes/primary.go
// Deskripsi : Executor untuk restore primary database dengan companion
// Author : Hadiyatna Muflihun
// Tanggal : 17 Desember 2025
// Last Modified : 5 Januari 2026
package modes

import (
	"context"
	"fmt"
	restoremodel "sfDBTools/internal/app/restore/model"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/ui"
	"strings"
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
func (e *PrimaryExecutor) Execute(ctx context.Context) (*restoremodel.RestoreResult, error) {
	startTime := time.Now()
	opts := e.service.GetPrimaryOptions()

	result := &restoremodel.RestoreResult{
		TargetDB:   opts.TargetDB,
		SourceFile: opts.File,
	}

	logger := e.service.GetLogger()
	logger.Info("Memulai proses restore database primary")
	e.service.SetRestoreInProgress(opts.TargetDB)
	defer e.service.ClearRestoreInProgress()

	// Dry-run mode: validasi file tanpa restore
	if opts.DryRun {
		logger.Info("Mode DRY-RUN: Validasi file tanpa restore...")
		return e.executeDryRun(ctx, opts, result, startTime)
	}

	// 2. Detect companion file if needed
	if opts.IncludeDmart {
		// Companion file sudah harus di-resolve saat setup (sebelum konfirmasi).
		// Executor tidak boleh prompt interaktif.
		if opts.CompanionFile != "" {
			result.CompanionFile = opts.CompanionFile
			result.CompanionDB = opts.TargetDB + consts.SuffixDmart
		} else {
			logger.Warn("Companion (dmart) diaktifkan tapi file tidak tersedia; skip restore companion")
			ui.PrintWarning("⚠️  Companion (dmart) diaktifkan tapi file tidak tersedia; skip restore companion")
		}
	}

	// 3. Execute common restore flow for primary database
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

	// 4. Handle companion database if requested
	if opts.IncludeDmart && opts.CompanionFile != "" {
		if err := e.restoreCompanionDatabase(ctx, opts, result); err != nil {
			if opts.StopOnError {
				result.Error = err
				return result, err
			}
			logger.Warnf("Gagal restore companion: %v", err)
		}
	}

	// 5. Restore user grants if available
	result.GrantsFile = opts.GrantsFile
	result.GrantsRestored = performGrantsRestore(ctx, e.service, opts.GrantsFile, false)

	// 6. Post-restore operations
	performPostRestoreOperations(ctx, e.service, opts.TargetDB)

	// Copy grants from primary to companion if available
	if opts.IncludeDmart && strings.TrimSpace(opts.CompanionFile) != "" {
		companionDB := opts.TargetDB + consts.SuffixDmart
		copyGrantsBetweenDatabases(ctx, e.service, opts.TargetDB, companionDB)
	}

	finalizeResult(result, startTime, true)
	logger.Info("Restore primary database berhasil")

	return result, nil
}

// restoreCompanionDatabase handles backup, drop, and restore for companion database
func (e *PrimaryExecutor) restoreCompanionDatabase(ctx context.Context, opts *restoremodel.RestorePrimaryOptions, result *restoremodel.RestoreResult) error {
	companionDB := opts.TargetDB + consts.SuffixDmart
	result.CompanionFile = opts.CompanionFile
	result.CompanionDB = companionDB

	// Use companion restore flow (handles backup, drop, and restore)
	flow := &companionRestoreFlow{
		service:       e.service,
		ctx:           ctx,
		primaryDB:     opts.TargetDB,
		sourceFile:    opts.CompanionFile,
		encryptionKey: opts.EncryptionKey,
		skipBackup:    opts.SkipBackup,
		dropTarget:    opts.DropTarget,
		stopOnError:   opts.StopOnError,
		backupOpts:    opts.BackupOptions,
	}

	backupFile, err := flow.execute()
	if err != nil {
		ui.PrintWarning(fmt.Sprintf("⚠️  Gagal restore companion database %s: %v", companionDB, err))
		return err
	}

	if backupFile != "" {
		result.CompanionBackup = backupFile
	}

	return nil
}

// executeDryRun melakukan validasi file backup tanpa restore
func (e *PrimaryExecutor) executeDryRun(ctx context.Context, opts *restoremodel.RestorePrimaryOptions, result *restoremodel.RestoreResult, startTime time.Time) (*restoremodel.RestoreResult, error) {
	validator := newDryRunValidator(e.service, ctx, result, startTime)

	// Validate primary file
	if err := validator.validateSingleFile(opts.File, opts.EncryptionKey); err != nil {
		result.Error = fmt.Errorf("gagal membuka file primary: %w", err)
		return result, result.Error
	}

	// Validate companion file if provided
	var companionValid bool
	if opts.IncludeDmart && opts.CompanionFile != "" {
		if err := validator.validateSingleFile(opts.CompanionFile, opts.EncryptionKey); err == nil {
			companionValid = true
			result.CompanionFile = opts.CompanionFile
			result.CompanionDB = opts.TargetDB + consts.SuffixDmart
		}
	}

	// Check database status
	dbExists, err := validator.validateDatabaseStatus(opts.TargetDB)
	if err != nil {
		result.Error = err
		return result, err
	}

	var companionExists bool
	if opts.IncludeDmart {
		companionDB := opts.TargetDB + consts.SuffixDmart
		companionExists, _ = e.service.GetTargetClient().CheckDatabaseExists(ctx, companionDB)
	}

	// Build summary info
	info := map[string]string{
		"Source File": opts.File,
		"Target DB":   opts.TargetDB,
		"DB Exists":   fmt.Sprintf("%v", dbExists),
	}
	if opts.IncludeDmart {
		if companionValid {
			info["Companion File"] = opts.CompanionFile
			info["Companion DB"] = fmt.Sprintf("%s (Exists: %v)", result.CompanionDB, companionExists)
		}
	}
	if dbExists && !opts.SkipBackup {
		info["Pre-restore Backup"] = "Will be created"
	}

	warnings := []string{}
	if !companionValid && opts.IncludeDmart {
		warnings = append(warnings, "Companion File: Not detected or invalid")
	}
	if opts.DropTarget && dbExists {
		warnings = append(warnings, "⚠️  Database will be DROPPED before restore")
	}

	validator.printSummary(info, warnings)
	return validator.finalize()
}
