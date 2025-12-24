package modes

import (
	"context"
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/ui"
	"strings"
	"time"
)

// SecondaryExecutor implements restore for secondary database.
// Source can be a backup file or a freshly generated backup from primary database.
type SecondaryExecutor struct {
	svc RestoreService
}

func NewSecondaryExecutor(svc RestoreService) *SecondaryExecutor {
	return &SecondaryExecutor{svc: svc}
}

func (e *SecondaryExecutor) Execute(ctx context.Context) (*types.RestoreResult, error) {
	startTime := time.Now()
	opts := e.svc.GetSecondaryOptions()
	logger := e.svc.GetLogger()

	result := &types.RestoreResult{
		TargetDB: opts.TargetDB,
	}

	logger.Info("Memulai proses restore database secondary")
	e.svc.SetRestoreInProgress(opts.TargetDB)
	defer e.svc.ClearRestoreInProgress()

	// Resolve source file(s)
	sourceFile := ""
	companionSourceFile := ""
	if opts.From == "primary" {
		// 1) Ensure primary exists
		exists, err := e.svc.GetTargetClient().CheckDatabaseExists(ctx, opts.PrimaryDB)
		if err != nil {
			result.Error = fmt.Errorf("gagal mengecek database primary: %w", err)
			return result, result.Error
		}
		if !exists {
			result.Error = fmt.Errorf("database primary tidak ditemukan: %s", opts.PrimaryDB)
			return result, result.Error
		}

		// 2) Backup primary (always)
		b, err := e.svc.BackupDatabaseIfNeeded(ctx, opts.PrimaryDB, true, false, opts.BackupOptions)
		if err != nil {
			result.Error = fmt.Errorf("gagal backup database primary: %w", err)
			return result, result.Error
		}
		sourceFile = b
		result.SourceFile = b

		// 2b) Backup primary companion (_dmart) jika diminta dan ada
		if opts.IncludeDmart {
			primaryDmart := opts.PrimaryDB + consts.SuffixDmart
			dmartExists, derr := e.svc.GetTargetClient().CheckDatabaseExists(ctx, primaryDmart)
			if derr != nil {
				result.Error = fmt.Errorf("gagal mengecek companion (dmart) database primary: %w", derr)
				return result, result.Error
			}
			if dmartExists {
				db, berr := e.svc.BackupDatabaseIfNeeded(ctx, primaryDmart, true, false, opts.BackupOptions)
				if berr != nil {
					if opts.StopOnError {
						result.Error = fmt.Errorf("gagal backup companion (dmart) database primary: %w", berr)
						return result, result.Error
					}
					ui.PrintWarning(fmt.Sprintf("⚠️  Gagal backup primary dmart (%s): %v", primaryDmart, berr))
				} else {
					companionSourceFile = db
					result.CompanionFile = db
					result.CompanionDB = opts.TargetDB + consts.SuffixDmart
				}
			} else {
				ui.PrintWarning(fmt.Sprintf("⚠️  Companion (dmart) primary tidak ditemukan: %s (skip)", primaryDmart))
			}
		}
	} else {
		sourceFile = opts.File
		result.SourceFile = opts.File
		if opts.IncludeDmart {
			companionSourceFile = opts.CompanionFile
			if strings.TrimSpace(companionSourceFile) != "" {
				result.CompanionFile = companionSourceFile
				result.CompanionDB = opts.TargetDB + consts.SuffixDmart
			}
		}
	}

	if sourceFile == "" {
		result.Error = fmt.Errorf("source file kosong (from=%s)", opts.From)
		return result, result.Error
	}

	// Dry-run: validate inputs only (no restore)
	if opts.DryRun {
		logger.Info("Mode DRY-RUN: validasi file tanpa restore...")
		ui.PrintInfo(fmt.Sprintf("  Source File: %s", sourceFile))
		ui.PrintInfo(fmt.Sprintf("  Target DB: %s", opts.TargetDB))
		if opts.IncludeDmart {
			if strings.TrimSpace(companionSourceFile) != "" {
				ui.PrintInfo(fmt.Sprintf("  Companion File: %s", companionSourceFile))
				ui.PrintInfo(fmt.Sprintf("  Companion DB: %s", opts.TargetDB+consts.SuffixDmart))
			} else {
				ui.PrintWarning("  Companion (dmart): not available / skipped")
			}
		}
		result.Success = true
		result.Duration = time.Since(startTime).Round(time.Second).String()
		return result, nil
	}

	// 3) Check target secondary existence
	targetExists, err := e.svc.GetTargetClient().CheckDatabaseExists(ctx, opts.TargetDB)
	if err != nil {
		result.Error = fmt.Errorf("gagal mengecek database secondary: %w", err)
		return result, result.Error
	}

	// 4) Backup target secondary if requested
	backupFile, err := e.svc.BackupDatabaseIfNeeded(ctx, opts.TargetDB, targetExists, opts.SkipBackup, opts.BackupOptions)
	if err != nil {
		result.Error = err
		return result, result.Error
	}
	result.BackupFile = backupFile

	// 5) Drop target secondary if requested
	if err := e.svc.DropDatabaseIfNeeded(ctx, opts.TargetDB, targetExists, opts.DropTarget); err != nil {
		result.Error = err
		return result, result.Error
	}
	if opts.DropTarget && targetExists {
		result.DroppedDB = true
	}

	// 6) Create and restore to secondary
	if err := e.svc.CreateAndRestoreDatabase(ctx, opts.TargetDB, sourceFile, opts.EncryptionKey); err != nil {
		result.Error = err
		return result, result.Error
	}

	// 7) Restore companion (dmart) if available
	if opts.IncludeDmart && strings.TrimSpace(companionSourceFile) != "" {
		companionDB := opts.TargetDB + consts.SuffixDmart
		companionExists, cerr := e.svc.GetTargetClient().CheckDatabaseExists(ctx, companionDB)
		if cerr != nil {
			result.Error = fmt.Errorf("gagal mengecek companion (dmart) target: %w", cerr)
			return result, result.Error
		}

		// Backup companion target if requested
		companionBackup, berr := e.svc.BackupDatabaseIfNeeded(ctx, companionDB, companionExists, opts.SkipBackup, opts.BackupOptions)
		if berr != nil {
			if opts.StopOnError {
				result.Error = fmt.Errorf("gagal backup companion (dmart) target: %w", berr)
				return result, result.Error
			}
			ui.PrintWarning(fmt.Sprintf("⚠️  Gagal backup target dmart (%s): %v", companionDB, berr))
		} else if companionBackup != "" {
			result.CompanionBackup = companionBackup
		}

		// Drop companion target if requested
		if err := e.svc.DropDatabaseIfNeeded(ctx, companionDB, companionExists, opts.DropTarget); err != nil {
			if opts.StopOnError {
				result.Error = fmt.Errorf("gagal drop companion (dmart) target: %w", err)
				return result, result.Error
			}
			ui.PrintWarning(fmt.Sprintf("⚠️  Gagal drop target dmart (%s): %v", companionDB, err))
		} else if opts.DropTarget && companionExists {
			result.DroppedCompanion = true
		}

		if err := e.svc.CreateAndRestoreDatabase(ctx, companionDB, companionSourceFile, opts.EncryptionKey); err != nil {
			if opts.StopOnError {
				result.Error = fmt.Errorf("gagal restore companion (dmart) target: %w", err)
				return result, result.Error
			}
			ui.PrintWarning(fmt.Sprintf("⚠️  Gagal restore companion (dmart) %s: %v", companionDB, err))
		}
	}

	// 8) Post-restore: copy grants (warning-only)
	// Khusus: jika source dari primary, copy grants primary -> secondary.
	if opts.From == "primary" {
		if gerr := e.svc.CopyDatabaseGrants(ctx, opts.PrimaryDB, opts.TargetDB); gerr != nil {
			logger.Warnf("Gagal copy grants primary -> secondary: %v", gerr)
			ui.PrintWarning(fmt.Sprintf("⚠️  Restore berhasil, tapi gagal copy grants primary -> secondary: %v", gerr))
		}
	}

	// Copy grants secondary -> secondary_dmart (jika dmart tersedia)
	if opts.IncludeDmart && strings.TrimSpace(companionSourceFile) != "" {
		companionDB := opts.TargetDB + consts.SuffixDmart
		if gerr := e.svc.CopyDatabaseGrants(ctx, opts.TargetDB, companionDB); gerr != nil {
			logger.Warnf("Gagal copy grants secondary -> secondary_dmart: %v", gerr)
			ui.PrintWarning(fmt.Sprintf("⚠️  Restore berhasil, tapi gagal copy grants secondary -> secondary_dmart: %v", gerr))
		}
	}

	// 9) Post-restore: buat database _temp (warning-only) + copy grants ke temp (warning-only)
	if !strings.HasSuffix(opts.TargetDB, consts.SuffixDmart) {
		tempDB, terr := e.svc.CreateTempDatabaseIfNeeded(ctx, opts.TargetDB)
		if terr != nil {
			logger.Warnf("Gagal membuat temp DB: %v", terr)
			ui.PrintWarning(fmt.Sprintf("⚠️  Restore berhasil, tapi gagal membuat temp DB: %v", terr))
		} else if strings.TrimSpace(tempDB) != "" {
			if gerr := e.svc.CopyDatabaseGrants(ctx, opts.TargetDB, tempDB); gerr != nil {
				logger.Warnf("Gagal copy grants ke temp DB: %v", gerr)
				ui.PrintWarning(fmt.Sprintf("⚠️  Restore berhasil, tapi gagal copy grants ke temp DB: %v", gerr))
			}
		}
	}

	result.Success = true
	result.Duration = time.Since(startTime).Round(time.Second).String()
	logger.Info("Restore secondary database berhasil")
	return result, nil
}
