// File : internal/restore/modes/dryrun_helpers.go
// Deskripsi : Dry-run validation helpers untuk semua restore executors
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 5 Januari 2026
package modes

import (
	"context"
	"fmt"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/ui/print"
	"time"
)

// dryRunValidator menangani validasi dry-run untuk operasi restore
type dryRunValidator struct {
	service   RestoreService
	ctx       context.Context
	result    *restoremodel.RestoreResult
	startTime time.Time
}

// newDryRunValidator membuat dry-run validator baru
func newDryRunValidator(service RestoreService, ctx context.Context, result *restoremodel.RestoreResult, startTime time.Time) *dryRunValidator {
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
	print.PrintSuccess("\n✓ Validasi File Backup:")
	for key, value := range info {
		print.PrintInfo(fmt.Sprintf("  %s: %s", key, value))
	}
	for _, warning := range warnings {
		print.PrintWarning(fmt.Sprintf("  %s", warning))
	}
}

// finalize mengatur result sebagai sukses dan mengembalikannya
func (v *dryRunValidator) finalize() (*restoremodel.RestoreResult, error) {
	finalizeResult(v.result, v.startTime, true)
	return v.result, nil
}
