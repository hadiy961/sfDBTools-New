// File : internal/restore/modes/primary.go
// Deskripsi : Executor untuk restore primary database dengan companion
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package modes

import (
	"context"
	"fmt"
	"sfDBTools/internal/restore/helpers"
	"sfDBTools/internal/types"
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
func (e *PrimaryExecutor) Execute(ctx context.Context) (*types.RestoreResult, error) {
	startTime := time.Now()
	opts := e.service.GetPrimaryOptions()

	result := &types.RestoreResult{
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

	// 1. Check if database exists
	dbExists, err := e.service.GetTargetClient().CheckDatabaseExists(ctx, opts.TargetDB)
	if err != nil {
		result.Error = fmt.Errorf("gagal mengecek database target: %w", err)
		return result, result.Error
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

	// 3. Backup primary database if needed
	backupFile, err := e.service.BackupDatabaseIfNeeded(ctx, opts.TargetDB, dbExists, opts.SkipBackup, opts.BackupOptions)
	if err != nil {
		result.Error = err
		return result, result.Error
	}
	result.BackupFile = backupFile

	// Backup companion if needed
	if opts.IncludeDmart && !opts.SkipBackup {
		companionDB := opts.TargetDB + consts.SuffixDmart
		companionExists, err := e.service.GetTargetClient().CheckDatabaseExists(ctx, companionDB)
		if err == nil && companionExists {
			companionBackup, err := e.service.BackupDatabaseIfNeeded(ctx, companionDB, true, false, opts.BackupOptions)
			if err != nil {
				logger.Warnf("Gagal backup companion database: %v", err)
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
		companionDB := opts.TargetDB + consts.SuffixDmart
		companionExists, err := e.service.GetTargetClient().CheckDatabaseExists(ctx, companionDB)
		if err == nil && companionExists {
			if err := e.service.DropDatabaseIfNeeded(ctx, companionDB, true, true); err != nil {
				logger.Warnf("Gagal drop companion database: %v", err)
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
		companionDB := opts.TargetDB + consts.SuffixDmart
		logger.Infof("Restore companion database dari %s...", opts.CompanionFile)

		if err := e.service.CreateAndRestoreDatabase(ctx, companionDB, opts.CompanionFile, opts.EncryptionKey); err != nil {
			if opts.StopOnError {
				result.Error = fmt.Errorf("gagal restore companion database %s: %w", companionDB, err)
				return result, result.Error
			}
			logger.Warnf("Gagal restore companion database: %v", err)
			ui.PrintWarning(fmt.Sprintf("⚠️  Gagal restore companion database %s: %v", companionDB, err))
		} else {
			logger.Infof("Companion database %s berhasil di-restore", companionDB)
		}
	}

	// 7. Restore user grants if available
	result.GrantsFile = opts.GrantsFile
	grantsRestored, err := e.service.RestoreUserGrantsIfAvailable(ctx, opts.GrantsFile)
	if err != nil {
		logger.Errorf("Gagal restore user grants: %v", err)
		ui.PrintWarning(fmt.Sprintf("⚠️  Database berhasil di-restore, tapi gagal restore user grants: %v", err))
		result.GrantsRestored = false
	} else {
		result.GrantsRestored = grantsRestored
	}

	// 8. Post-restore: buat database _temp (warning-only) + copy grants (warning-only)
	if !strings.HasSuffix(opts.TargetDB, consts.SuffixDmart) {
		tempDB, terr := e.service.CreateTempDatabaseIfNeeded(ctx, opts.TargetDB)
		if terr != nil {
			logger.Warnf("Gagal membuat temp DB: %v", terr)
			ui.PrintWarning(fmt.Sprintf("⚠️  Restore berhasil, tapi gagal membuat temp DB: %v", terr))
		} else if strings.TrimSpace(tempDB) != "" {
			if gerr := e.service.CopyDatabaseGrants(ctx, opts.TargetDB, tempDB); gerr != nil {
				logger.Warnf("Gagal copy grants ke temp DB: %v", gerr)
				ui.PrintWarning(fmt.Sprintf("⚠️  Restore berhasil, tapi gagal copy grants ke temp DB: %v", gerr))
			}
		}
	}

	// Copy grants dari primary ke primary_dmart (jika dmart di-restore / ada)
	if opts.IncludeDmart && strings.TrimSpace(opts.CompanionFile) != "" {
		companionDB := opts.TargetDB + consts.SuffixDmart
		if gerr := e.service.CopyDatabaseGrants(ctx, opts.TargetDB, companionDB); gerr != nil {
			logger.Warnf("Gagal copy grants ke companion DB: %v", gerr)
			ui.PrintWarning(fmt.Sprintf("⚠️  Restore berhasil, tapi gagal copy grants ke companion DB: %v", gerr))
		}
	}

	result.Success = true
	result.Duration = time.Since(startTime).Round(time.Second).String()
	logger.Info("Restore primary database berhasil")

	return result, nil
}

// executeDryRun melakukan validasi file backup tanpa restore
func (e *PrimaryExecutor) executeDryRun(ctx context.Context, opts *types.RestorePrimaryOptions, result *types.RestoreResult, startTime time.Time) (*types.RestoreResult, error) {
	logger := e.service.GetLogger()
	logger.Info("Validasi file backup primary...")

	// Validasi primary file
	reader, closers, err := helpers.OpenAndPrepareReader(opts.File, opts.EncryptionKey)
	if err != nil {
		result.Error = fmt.Errorf("gagal membuka file primary: %w", err)
		return result, result.Error
	}
	helpers.CloseReaders(closers)
	_ = reader

	// Validasi companion file jika ada
	var companionValid bool
	if opts.IncludeDmart && opts.CompanionFile != "" {
		reader, closers, err := helpers.OpenAndPrepareReader(opts.CompanionFile, opts.EncryptionKey)
		if err == nil {
			helpers.CloseReaders(closers)
			_ = reader
			companionValid = true
			result.CompanionFile = opts.CompanionFile
			result.CompanionDB = opts.TargetDB + consts.SuffixDmart
		}
	}

	// Check database status
	dbExists, err := e.service.GetTargetClient().CheckDatabaseExists(ctx, opts.TargetDB)
	if err != nil {
		result.Error = fmt.Errorf("gagal mengecek database target: %w", err)
		return result, result.Error
	}

	var companionExists bool
	if opts.IncludeDmart {
		companionDB := opts.TargetDB + consts.SuffixDmart
		companionExists, _ = e.service.GetTargetClient().CheckDatabaseExists(ctx, companionDB)
	}

	// Print hasil validasi
	ui.PrintSuccess("\n✓ Validasi File Backup Primary:")
	ui.PrintInfo(fmt.Sprintf("  Source File: %s", opts.File))
	ui.PrintInfo(fmt.Sprintf("  Target DB: %s", opts.TargetDB))
	ui.PrintInfo(fmt.Sprintf("  DB Exists: %v", dbExists))
	if opts.IncludeDmart {
		if companionValid {
			ui.PrintInfo(fmt.Sprintf("  Companion File: %s", opts.CompanionFile))
			ui.PrintInfo(fmt.Sprintf("  Companion DB: %s (Exists: %v)", result.CompanionDB, companionExists))
		} else {
			ui.PrintWarning("  Companion File: Not detected or invalid")
		}
	}
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
