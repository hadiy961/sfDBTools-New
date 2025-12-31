// File : internal/backup/execution/loop.go
// Deskripsi : Loop execution logic untuk multi-database backup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-31
// Last Modified : 2025-12-31

package execution

import (
	"context"
	"fmt"

	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/consts"
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

	for idx, dbName := range databases {
		// Check context cancellation
		if ctx.Err() != nil {
			e.Log.Warn("Proses backup dibatalkan")
			result.Errors = append(result.Errors, "Backup dibatalkan oleh user")
			break
		}

		e.executeSingleBackupInLoop(ctx, dbName, idx+1, len(databases), config, outputPathFunc, &result)
	}

	return result
}

// executeSingleBackupInLoop executes backup untuk satu database dalam loop context.
// Updates result object dengan success/failure info.
func (e *Engine) executeSingleBackupInLoop(
	ctx context.Context,
	dbName string,
	currentIdx, totalDBs int,
	config types_backup.BackupLoopConfig,
	outputPathFunc func(string) (string, error),
	result *types_backup.BackupLoopResult,
) {
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
		return
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
		result.FailedDBs = append(result.FailedDBs, types_backup.FailedDatabaseInfo{
			DatabaseName: dbName,
			Error:        err.Error(),
		})
		result.Failed++
		return
	}

	result.BackupInfos = append(result.BackupInfos, backupInfo)
	result.Success++

	// Export user grants untuk separated/single modes
	if e.UserGrants != nil {
		if config.Mode == consts.ModeSeparated || config.Mode == consts.ModeSingle {
			path := e.UserGrants.ExportUserGrantsIfNeeded(ctx, outputPath, []string{dbName})
			if e.Config.Backup.Output.SaveBackupInfo {
				e.UserGrants.UpdateMetadataUserGrantsPath(outputPath, path)
			}
		}
	}
}
