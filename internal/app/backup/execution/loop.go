// File : internal/backup/execution/loop.go
// Deskripsi : Loop execution logic untuk multi-database backup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-31
// Last Modified : 2026-01-02

package execution

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"sfdbtools/internal/app/backup/model/types_backup"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
)

// ExecuteBackupLoop menjalankan backup across multiple databases.
// Menggunakan ExecuteAndBuildBackup untuk setiap database.
func (e *Engine) ExecuteBackupLoop(
	ctx context.Context,
	databases []string,
	config types_backup.BackupLoopConfig,
	outputPathFunc func(dbName string) (string, error),
) types_backup.BackupLoopResult {
	result := types_backup.BackupLoopResult{
		BackupInfos: make([]types_backup.DatabaseBackupInfo, 0),
		FailedDBs:   make([]types_backup.FailedDatabaseInfo, 0),
		Errors:      make([]string, 0),
	}

	if len(databases) == 0 {
		e.Log.Warn("Tidak ada database yang dipilih untuk backup")
		result.Errors = append(result.Errors, "tidak ada database yang dipilih")
		return result
	}

	// Register cleanup handler untuk context cancellation
	// Cleanup akan dijalankan jika user cancel backup (CTRL+C)
	// Type assertion untuk mengakses concrete type BackupExecutionState
	type cleanupState interface {
		StateTracker
		EnableCleanup(log applog.Logger)
	}
	if cs, ok := e.State.(cleanupState); ok && cs != nil {
		cs.EnableCleanup(e.Log)

		// IMPORTANT:
		// Jangan gunakan context.WithCancel(ctx) + defer cancel() untuk menghentikan goroutine,
		// karena itu akan memicu Err()==context.Canceled saat function normal selesai.
		// Akibatnya log "Backup cancelled..." muncul walaupun backup berhasil.
		stopCleanup := make(chan struct{})
		defer close(stopCleanup)
		go func() {
			select {
			case <-ctx.Done():
				// Hanya cleanup saat upstream context memang dibatalkan (CTRL+C / SIGTERM / timeout)
				if errors.Is(ctx.Err(), context.Canceled) {
					e.Log.Warn("⚠️  Backup cancelled, cleaning up partial files...")
					cs.Cleanup()
				}
			case <-stopCleanup:
				// Backup loop selesai normal, hentikan goroutine tanpa menjalankan cleanup
				return
			}
		}()
	}

	for idx, dbName := range databases {
		// Check context cancellation
		if ctx.Err() != nil {
			e.Log.Warn("Proses backup dibatalkan")
			result.Errors = append(result.Errors, "Backup dibatalkan oleh user")
			break
		}

		if err := e.executeSingleBackupInLoop(ctx, dbName, idx+1, len(databases), config, outputPathFunc, &result); err != nil {
			e.Log.Errorf("Menghentikan proses backup karena error kritikal: %v", err)
			result.Errors = append(result.Errors, "Proses backup dihentikan karena error koneksi/autentikasi fatal")
			break
		}
	}

	return result
}

// executeSingleBackupInLoop executes backup untuk satu database dalam loop context.
// Updates result object dengan success/failure info.
// Mengembalikan error non-nil jika terjadi error fatal (connection/auth) untuk menghentikan loop.
func (e *Engine) executeSingleBackupInLoop(
	ctx context.Context,
	dbName string,
	currentIdx, totalDBs int,
	config types_backup.BackupLoopConfig,
	outputPathFunc func(string) (string, error),
	result *types_backup.BackupLoopResult,
) error {
	// Early context check BEFORE starting backup
	// Mencegah partial backup jika context sudah cancelled
	select {
	case <-ctx.Done():
		e.Log.Warnf("⚠️  Backup cancelled before database %s (context done)", dbName)
		result.Errors = append(result.Errors, fmt.Sprintf("Backup cancelled before %s", dbName))
		return ctx.Err()
	default:
		// Context masih aktif, lanjut backup
	}

	start := time.Now()
	e.Log.Infof("[%d/%d] Backup database: %s", currentIdx, totalDBs, dbName)

	// Generate output path untuk database ini
	outputPath, err := outputPathFunc(dbName)
	if err != nil {
		msg := fmt.Sprintf("gagal generate path untuk %s: %v", dbName, err)
		e.Log.Error(msg)
		result.FailedDBs = append(result.FailedDBs, types_backup.FailedDatabaseInfo{
			DatabaseName: dbName,
			Error:        msg,
		})
		result.Failed++
		return nil
	}

	// Execute backup untuk database ini
	backupInfo, err := e.ExecuteAndBuildBackup(ctx, types_backup.BackupExecutionConfig{
		DBName:       dbName,
		OutputPath:   outputPath,
		BackupType:   config.BackupType,
		TotalDBFound: config.TotalDBs,
		IsMultiDB:    false,
	})

	if err != nil {
		e.Log.Warnf("[%d/%d] Backup database gagal: %s (%s)", currentIdx, totalDBs, dbName, time.Since(start).Round(time.Millisecond))
		result.FailedDBs = append(result.FailedDBs, types_backup.FailedDatabaseInfo{
			DatabaseName: dbName,
			Error:        err.Error(),
		})
		result.Failed++

		// Circuit Breaker: jika error terkait koneksi atau autentikasi yang fatal, hentikan loop
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "access denied") ||
			strings.Contains(errStr, "connection refused") ||
			strings.Contains(errStr, "can't connect") ||
			strings.Contains(errStr, "unknown server") {
			return fmt.Errorf("fatal connection/auth error: %w", err)
		}
		return nil
	}

	result.BackupInfos = append(result.BackupInfos, backupInfo)
	result.Success++
	e.Log.Infof("[%d/%d] Selesai backup database: %s (%s)", currentIdx, totalDBs, dbName, time.Since(start).Round(time.Millisecond))

	// Export user grants untuk separated/single modes
	if e.UserGrants != nil {
		if config.Mode == consts.ModeSeparated || config.Mode == consts.ModeSingle {
			userDefPath, userGrantsPath, ugErr := e.UserGrants.ExportUserGrantsIfNeeded(ctx, outputPath, []string{dbName}, e.Options.ExcludeGrant)
			if ugErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("export user grants gagal: %s", ugErr.Error()))
			}
			if e.Config.Backup.Output.SaveBackupInfo {
				permissions := e.Config.Backup.Output.MetadataPermissions
				e.UserGrants.UpdateMetadataUserGrantsPath(outputPath, userDefPath, userGrantsPath, permissions)
			}
		}
	}

	return nil
}
