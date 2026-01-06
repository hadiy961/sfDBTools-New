// File : internal/restore/modes/common_helpers.go
// Deskripsi : Common helper functions untuk semua restore executors
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 5 Januari 2026
package modes

import (
	"context"
	"fmt"
	"sfdbtools/internal/app/restore/helpers"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/ui/print"
	"sfdbtools/pkg/consts"
	"strings"
	"time"
)

// commonRestoreFlow menjalankan alur standar backup -> drop -> restore untuk single database
type commonRestoreFlow struct {
	service       RestoreService
	ctx           context.Context
	dbName        string
	sourceFile    string
	encryptionKey string
	skipBackup    bool
	dropTarget    bool
	stopOnError   bool
	backupOpts    *restoremodel.RestoreBackupOptions
}

// execute menjalankan alur restore umum dan mengembalikan path file backup serta error
func (f *commonRestoreFlow) execute() (backupFile string, err error) {
	client := f.service.GetTargetClient()

	// 1. Cek apakah database sudah ada
	dbExists, err := client.CheckDatabaseExists(f.ctx, f.dbName)
	if err != nil {
		return "", fmt.Errorf("gagal mengecek database %s: %w", f.dbName, err)
	}

	// 2. Backup database jika diperlukan
	backupFile, err = f.service.BackupDatabaseIfNeeded(f.ctx, f.dbName, dbExists, f.skipBackup, f.backupOpts)
	if err != nil {
		return "", err
	}

	// 3. Drop database jika diperlukan
	if err := f.service.DropDatabaseIfNeeded(f.ctx, f.dbName, dbExists, f.dropTarget); err != nil {
		return backupFile, err
	}

	// 4. Create and restore database
	if err := f.service.CreateAndRestoreDatabase(f.ctx, f.dbName, f.sourceFile, f.encryptionKey); err != nil {
		return backupFile, err
	}

	return backupFile, nil
}

// performPostRestoreOperations menangani pembuatan temp DB dan copy grants (failure hanya warning)
func performPostRestoreOperations(ctx context.Context, service RestoreService, primaryDB string) {
	logger := service.GetLogger()

	// Skip jika ini adalah database dmart
	if strings.HasSuffix(primaryDB, consts.SuffixDmart) {
		return
	}

	// Create temp database
	tempDB, err := service.CreateTempDatabaseIfNeeded(ctx, primaryDB)
	if err != nil {
		logger.Warnf("Gagal membuat temp DB: %v", err)
		print.PrintWarning(fmt.Sprintf("⚠️  Restore berhasil, tapi gagal membuat temp DB: %v", err))
		return
	}

	if strings.TrimSpace(tempDB) == "" {
		return
	}

	// Copy grants ke temp database
	if err := service.CopyDatabaseGrants(ctx, primaryDB, tempDB); err != nil {
		logger.Warnf("Gagal copy grants ke temp DB: %v", err)
		print.PrintWarning(fmt.Sprintf("⚠️  Restore berhasil, tapi gagal copy grants ke temp DB: %v", err))
	}
}

// performGrantsRestore menangani restore user grants (failure hanya warning)
func performGrantsRestore(ctx context.Context, service RestoreService, grantsFile string, skipGrants bool) (grantsRestored bool) {
	if skipGrants {
		return false
	}

	logger := service.GetLogger()
	grantsRestored, err := service.RestoreUserGrantsIfAvailable(ctx, grantsFile)
	if err != nil {
		logger.Errorf("Gagal restore user grants: %v", err)
		print.PrintWarning(fmt.Sprintf("⚠️  Database berhasil di-restore, tapi gagal restore user grants: %v", err))
		return false
	}

	return grantsRestored
}

// copyGrantsBetweenDatabases menyalin grants dari source ke target database (failure hanya warning)
func copyGrantsBetweenDatabases(ctx context.Context, service RestoreService, sourceDB, targetDB string) {
	logger := service.GetLogger()
	if err := service.CopyDatabaseGrants(ctx, sourceDB, targetDB); err != nil {
		logger.Warnf("Gagal copy grants %s -> %s: %v", sourceDB, targetDB, err)
		print.PrintWarning(fmt.Sprintf("⚠️  Restore berhasil, tapi gagal copy grants %s -> %s: %v", sourceDB, targetDB, err))
	}
}

// validateFileForDryRun memvalidasi bahwa file backup bisa dibuka dan dibaca
func validateFileForDryRun(file, encryptionKey string) error {
	reader, closers, err := helpers.OpenAndPrepareReader(file, encryptionKey)
	if err != nil {
		return fmt.Errorf("gagal membuka file: %w", err)
	}
	defer helpers.CloseReaders(closers)
	_ = reader // File berhasil dibuka
	return nil
}

// createResultWithDefaults membuat RestoreResult dengan nilai default umum
func createResultWithDefaults(targetDB, sourceFile string, startTime time.Time) *restoremodel.RestoreResult {
	return &restoremodel.RestoreResult{
		TargetDB:   targetDB,
		SourceFile: sourceFile,
		Success:    false, // Akan di-set true jika berhasil
	}
}

// finalizeResult mengatur status sukses dan durasi untuk hasil restore
func finalizeResult(result *restoremodel.RestoreResult, startTime time.Time, success bool) {
	result.Success = success
	result.Duration = time.Since(startTime).Round(time.Second).String()
}
