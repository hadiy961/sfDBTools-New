package modes

import (
	"context"
	"fmt"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/ui/print"
	"sfdbtools/pkg/consts"
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

func (e *SecondaryExecutor) Execute(ctx context.Context) (*restoremodel.RestoreResult, error) {
	startTime := time.Now()
	opts := e.svc.GetSecondaryOptions()
	logger := e.svc.GetLogger()

	result := &restoremodel.RestoreResult{
		TargetDB: opts.TargetDB,
	}

	logger.Info("Memulai proses restore database secondary")
	e.svc.SetRestoreInProgress(opts.TargetDB)
	defer e.svc.ClearRestoreInProgress()

	// Resolve source file(s)
	resolver := newSourceFileResolver(e.svc, ctx, opts.From, opts.PrimaryDB, opts.File, opts.BackupOptions, opts.StopOnError)

	sourceFile, err := resolver.resolveMainSource()
	if err != nil {
		result.Error = err
		return result, err
	}
	result.SourceFile = sourceFile

	if err := validateSourceFile(sourceFile, opts.From); err != nil {
		result.Error = err
		return result, err
	}

	// Resolve companion source if needed
	var companionSourceFile string
	if opts.IncludeDmart {
		if opts.From == "primary" {
			companionSourceFile, err = resolver.backupCompanionIfExists()
			if err != nil {
				result.Error = err
				return result, err
			}
		} else {
			companionSourceFile = opts.CompanionFile
		}

		if strings.TrimSpace(companionSourceFile) != "" {
			result.CompanionFile = companionSourceFile
			result.CompanionDB = opts.TargetDB + consts.SuffixDmart
		}
	}

	// Dry-run: validate inputs only (no restore)
	if opts.DryRun {
		logger.Info("Mode DRY-RUN: validasi file tanpa restore...")
		print.PrintInfo(fmt.Sprintf("  Source File: %s", sourceFile))
		print.PrintInfo(fmt.Sprintf("  Target DB: %s", opts.TargetDB))
		if opts.IncludeDmart {
			if strings.TrimSpace(companionSourceFile) != "" {
				print.PrintInfo(fmt.Sprintf("  Companion File: %s", companionSourceFile))
				print.PrintInfo(fmt.Sprintf("  Companion DB: %s", opts.TargetDB+consts.SuffixDmart))
			} else {
				print.PrintWarning("  Companion (dmart): not available / skipped")
			}
		}
		result.Success = true
		result.Duration = time.Since(startTime).Round(time.Second).String()
		return result, nil
	}

	// Execute common restore flow for secondary database
	flow := &commonRestoreFlow{
		service:       e.svc,
		ctx:           ctx,
		dbName:        opts.TargetDB,
		sourceFile:    sourceFile,
		encryptionKey: opts.EncryptionKey,
		skipBackup:    opts.SkipBackup,
		dropTarget:    opts.DropTarget,
		stopOnError:   opts.StopOnError,
		backupOpts:    opts.BackupOptions,
	}

	backupFile, err := flow.execute()
	if err != nil {
		result.Error = err
		return result, err
	}
	result.BackupFile = backupFile

	// Restore companion database if available
	if opts.IncludeDmart && strings.TrimSpace(companionSourceFile) != "" {
		companionFlow := &companionRestoreFlow{
			service:       e.svc,
			ctx:           ctx,
			primaryDB:     opts.TargetDB,
			sourceFile:    companionSourceFile,
			encryptionKey: opts.EncryptionKey,
			skipBackup:    opts.SkipBackup,
			dropTarget:    opts.DropTarget,
			stopOnError:   opts.StopOnError,
			backupOpts:    opts.BackupOptions,
		}

		companionBackup, err := companionFlow.execute()
		if err != nil && opts.StopOnError {
			result.Error = err
			return result, err
		}
		if companionBackup != "" {
			result.CompanionBackup = companionBackup
		}
	}

	// Post-restore: copy grants from primary if applicable
	if opts.From == "primary" {
		copyGrantsBetweenDatabases(ctx, e.svc, opts.PrimaryDB, opts.TargetDB)
	}

	// Copy grants to companion if available
	if opts.IncludeDmart && strings.TrimSpace(companionSourceFile) != "" {
		companionDB := opts.TargetDB + consts.SuffixDmart
		copyGrantsBetweenDatabases(ctx, e.svc, opts.TargetDB, companionDB)
	}

	// Post-restore operations (temp DB creation)
	performPostRestoreOperations(ctx, e.svc, opts.TargetDB)

	finalizeResult(result, startTime, true)
	logger.Info("Restore secondary database berhasil")
	return result, nil
}
