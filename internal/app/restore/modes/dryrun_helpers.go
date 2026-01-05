// File : internal/restore/modes/dryrun_helpers.go
// Deskripsi : Dry-run validation helpers untuk semua restore executors
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package modes

import (
	"context"
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/ui"
	"time"
)

// dryRunValidator menangani validasi dry-run untuk operasi restore
type dryRunValidator struct {
	service   RestoreService
	ctx       context.Context
	result    *types.RestoreResult
	startTime time.Time
}

// newDryRunValidator membuat dry-run validator baru
func newDryRunValidator(service RestoreService, ctx context.Context, result *types.RestoreResult, startTime time.Time) *dryRunValidator {
	return &dryRunValidator{
		service:   service,
		ctx:       ctx,
		result:    result,
		startTime: startTime,
	}
}

// validateSingleFile memvalidasi single backup file untuk mode dry-run
func (v *dryRunValidator) validateSingleFile(file, encryptionKey string) error {
	if err := validateFileForDryRun(file, encryptionKey); err != nil {
		return err
	}
	return nil
}

// validateDatabaseStatus mengecek apakah target database ada
func (v *dryRunValidator) validateDatabaseStatus(dbName string) (bool, error) {
	client := v.service.GetTargetClient()
	dbExists, err := client.CheckDatabaseExists(v.ctx, dbName)
	if err != nil {
		return false, fmt.Errorf("gagal mengecek database target: %w", err)
	}
	return dbExists, nil
}

// printSummary mencetak ringkasan validasi dry-run
func (v *dryRunValidator) printSummary(info map[string]string, warnings []string) {
	ui.PrintSuccess("\n✓ Validasi File Backup:")
	for key, value := range info {
		ui.PrintInfo(fmt.Sprintf("  %s: %s", key, value))
	}
	for _, warning := range warnings {
		ui.PrintWarning(fmt.Sprintf("  %s", warning))
	}
}

// finalize mengatur result sebagai sukses dan mengembalikannya
func (v *dryRunValidator) finalize() (*types.RestoreResult, error) {
	finalizeResult(v.result, v.startTime, true)
	return v.result, nil
}
