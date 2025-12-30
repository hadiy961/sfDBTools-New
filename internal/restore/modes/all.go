// File : internal/restore/modes/all.go
// Deskripsi : Executor untuk restore all databases dengan streaming filtering
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-18
// Last Modified : 2025-12-30

package modes

import (
	"context"
	"sfDBTools/internal/types"
	"time"
)

// AllExecutor implements restore for all databases dengan streaming filtering
type AllExecutor struct {
	service RestoreService
}

// NewAllExecutor creates a new AllExecutor
func NewAllExecutor(svc RestoreService) *AllExecutor {
	return &AllExecutor{service: svc}
}

// Execute executes all databases restore dengan streaming processing
func (e *AllExecutor) Execute(ctx context.Context) (*types.RestoreResult, error) {
	startTime := time.Now()
	opts := e.service.GetAllOptions()

	result := &types.RestoreResult{
		TargetDB:   "ALL_DATABASES",
		SourceFile: opts.File,
		Success:    false,
	}

	logger := e.service.GetLogger()
	logger.Info("Memulai proses restore all databases")
	e.service.SetRestoreInProgress("ALL_DATABASES")
	defer e.service.ClearRestoreInProgress()

	// Dry run mode - hanya analisis file tanpa restore
	if opts.DryRun {
		logger.Info("Mode DRY-RUN: Analisis file dump tanpa restore...")
		return e.executeDryRun(ctx, opts, result)
	}

	// Execute actual restore dengan streaming
	if err := e.executeStreamingRestore(ctx, opts); err != nil {
		result.Error = err
		return result, err
	}

	// Restore user grants if available (optional)
	if !opts.SkipGrants {
		result.GrantsFile = opts.GrantsFile
		grantsRestored, err := e.service.RestoreUserGrantsIfAvailable(ctx, opts.GrantsFile)
		if err != nil {
			logger.Errorf("Gagal restore user grants: %v", err)
			result.GrantsRestored = false
		} else {
			result.GrantsRestored = grantsRestored
		}
	}

	result.Success = true
	result.Duration = time.Since(startTime).Round(time.Second).String()
	logger.Info("Restore all databases berhasil")

	return result, nil
}
