// File : internal/restore/modes/bulk_helpers.go
// Deskripsi : Helper functions untuk bulk/batch database operations
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 5 Januari 2026
package modes

import (
	"context"
	"fmt"
	restoremodel "sfDBTools/internal/app/restore/model"
	"strings"
	"time"
)

// bulkBackupHelper menangani backup multiple databases
type bulkBackupHelper struct {
	service     RestoreService
	ctx         context.Context
	databases   []string
	backupOpts  *restoremodel.RestoreBackupOptions
	stopOnError bool
}

// executeBackup melakukan backup untuk semua database dalam list
func (h *bulkBackupHelper) executeBackup() (backupFile string, err error) {
	logger := h.service.GetLogger()

	if len(h.databases) == 0 {
		logger.Info("Pre-restore backup: tidak ada database existing untuk dibackup")
		return "", nil
	}

	outDir := ""
	if h.backupOpts != nil {
		outDir = h.backupOpts.OutputDir
	}
	if outDir == "" {
		outDir = "(default dari config)"
	}

	logger.Infof("Pre-restore backup (single-file): membackup %d database ke %s", len(h.databases), outDir)
	backupStart := time.Now()
	backupFile, err = h.service.BackupDatabasesSingleFileIfNeeded(h.ctx, h.databases, false, h.backupOpts)
	backupDuration := time.Since(backupStart).Round(time.Millisecond)

	if err != nil {
		if h.stopOnError {
			return "", fmt.Errorf("backup pre-restore gagal: %w", err)
		}
		logger.Warnf("backup pre-restore gagal (durasi: %s): %v (lanjut)", backupDuration, err)
		return "", nil
	}

	if backupFile != "" {
		logger.Infof("Pre-restore backup selesai (durasi: %s): %s", backupDuration, backupFile)
	} else {
		logger.Infof("Pre-restore backup selesai (durasi: %s)", backupDuration)
	}

	return backupFile, nil
}

// bulkDropHelper menangani dropping multiple databases
type bulkDropHelper struct {
	service     RestoreService
	ctx         context.Context
	databases   []string
	stopOnError bool
}

// executeDrop melakukan drop untuk semua database dalam list
func (h *bulkDropHelper) executeDrop() (droppedCount int, err error) {
	logger := h.service.GetLogger()
	client := h.service.GetTargetClient()

	if len(h.databases) == 0 {
		logger.Info("Tidak ada database yang perlu di-drop")
		return 0, nil
	}

	logger.Infof("Database yang akan di-drop (%d): %s", len(h.databases), strings.Join(h.databases, ", "))

	dropStart := time.Now()
	droppedCount = 0

	for _, dbName := range h.databases {
		if dropErr := client.DropDatabase(h.ctx, dbName); dropErr != nil {
			if h.stopOnError {
				return droppedCount, fmt.Errorf("gagal drop database %s: %w", dbName, dropErr)
			}
			logger.Warnf("Gagal drop database %s: %v (lanjut)", dbName, dropErr)
			continue
		}
		droppedCount++
	}

	dropDuration := time.Since(dropStart).Round(time.Millisecond)
	logger.Infof("Berhasil drop %d/%d database (durasi: %s)", droppedCount, len(h.databases), dropDuration)

	return droppedCount, nil
}

// prepareTargetDatabases mengecek keberadaan database dan menyiapkan list backup/drop
func prepareTargetDatabases(ctx context.Context, service RestoreService, dbNames []string, skipBackup, dropTarget, stopOnError bool) (toBackup, toDrop []string, existsCount int, err error) {
	logger := service.GetLogger()
	client := service.GetTargetClient()

	toBackup = make([]string, 0, len(dbNames))
	toDrop = make([]string, 0, len(dbNames))
	existsCount = 0

	logger.Info("Memeriksa keberadaan database pada server target...")

	for _, dbName := range dbNames {
		exists, chkErr := client.CheckDatabaseExists(ctx, dbName)
		if chkErr != nil {
			if stopOnError {
				return nil, nil, 0, fmt.Errorf("gagal mengecek database %s: %w", dbName, chkErr)
			}
			logger.Warnf("Gagal cek database %s: %v (lanjut)", dbName, chkErr)
			continue
		}

		if exists {
			existsCount++
			if !skipBackup {
				toBackup = append(toBackup, dbName)
			}
			if dropTarget {
				toDrop = append(toDrop, dbName)
			}
		}
	}

	logger.Infof("Ringkasan pengecekan: %d dari %d database sudah ada", existsCount, len(dbNames))
	return toBackup, toDrop, existsCount, nil
}
