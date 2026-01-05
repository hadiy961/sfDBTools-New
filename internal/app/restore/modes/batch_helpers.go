// File : internal/restore/modes/batch_helpers.go
// Deskripsi : Helper functions untuk batch restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified :  2026-01-05
package modes

import (
	"context"
	"fmt"
	restoremodel "sfDBTools/internal/app/restore/model"
)

// batchRestoreTracker melacak jumlah sukses/gagal untuk operasi batch
type batchRestoreTracker struct {
	total    int
	success  int
	failures int
}

// newBatchRestoreTracker membuat batch tracker baru
func newBatchRestoreTracker(total int) *batchRestoreTracker {
	return &batchRestoreTracker{
		total:    total,
		success:  0,
		failures: 0,
	}
}

// recordSuccess menambah jumlah sukses
func (t *batchRestoreTracker) recordSuccess() {
	t.success++
}

// recordFailure menambah jumlah gagal
func (t *batchRestoreTracker) recordFailure() {
	t.failures++
}

// getSummary mengembalikan string ringkasan terformat
func (t *batchRestoreTracker) getSummary() string {
	return fmt.Sprintf("%d berhasil, %d gagal dari %d entri", t.success, t.failures, t.total)
}

// isAllSuccess mengembalikan true jika semua operasi sukses
func (t *batchRestoreTracker) isAllSuccess() bool {
	return t.failures == 0
}

// singleDatabaseRestore mengenkapsulasi parameter untuk restore single database dalam batch
type singleDatabaseRestore struct {
	service       RestoreService
	ctx           context.Context
	dbName        string
	sourceFile    string
	encryptionKey string
	grantsFile    string
	skipBackup    bool
	dropTarget    bool
	stopOnError   bool
	backupOpts    *restoremodel.RestoreBackupOptions
}

// execute melakukan restore untuk single database dalam mode batch
func (r *singleDatabaseRestore) execute() error {
	logger := r.service.GetLogger()
	client := r.service.GetTargetClient()

	// Cek keberadaan database
	dbExists, err := client.CheckDatabaseExists(r.ctx, r.dbName)
	if err != nil {
		return fmt.Errorf("cek database gagal: %w", err)
	}

	// Backup jika diperlukan
	if !r.skipBackup {
		if _, err := r.service.BackupDatabaseIfNeeded(r.ctx, r.dbName, dbExists, r.skipBackup, r.backupOpts); err != nil {
			return fmt.Errorf("backup pre-restore gagal: %w", err)
		}
	}

	// Drop jika diperlukan
	if err := r.service.DropDatabaseIfNeeded(r.ctx, r.dbName, dbExists, r.dropTarget); err != nil {
		return fmt.Errorf("drop database gagal: %w", err)
	}

	// Buat dan restore
	if err := r.service.CreateAndRestoreDatabase(r.ctx, r.dbName, r.sourceFile, r.encryptionKey); err != nil {
		return fmt.Errorf("restore gagal: %w", err)
	}

	// Restore grants jika tersedia (non-fatal)
	if r.grantsFile != "" {
		if _, err := r.service.RestoreUserGrantsIfAvailable(r.ctx, r.grantsFile); err != nil {
			if r.stopOnError {
				return fmt.Errorf("restore grants gagal: %w", err)
			}
			logger.Warnf("restore grants gagal: %v (lanjut)", err)
		}
	}

	// Operasi post-restore (pembuatan temp DB - non-fatal)
	performPostRestoreOperations(r.ctx, r.service, r.dbName)

	return nil
}
