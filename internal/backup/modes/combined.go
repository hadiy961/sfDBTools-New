// File : internal/backup/modes/combined.go
// Deskripsi : Mode backup combined - semua database dalam satu file
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package modes

import (
	"context"
	"fmt"
	"path/filepath"
	"sfDBTools/internal/types/types_backup"
	"strings"
)

// CombinedExecutor menangani backup combined mode
type CombinedExecutor struct {
	service BackupService
}

// NewCombinedExecutor membuat instance baru CombinedExecutor
func NewCombinedExecutor(svc BackupService) *CombinedExecutor {
	return &CombinedExecutor{service: svc}
}

// Execute melakukan backup semua database dalam satu file
func (e *CombinedExecutor) Execute(ctx context.Context, dbFiltered []string) types_backup.BackupResult {
	var res types_backup.BackupResult
	logger := e.service.GetLogger()
	logger.Info("Melakukan backup database dalam mode combined")

	totalDBFound := e.service.GetTotalDatabaseCount(ctx, dbFiltered)

	// Initialize result statistics
	res.TotalDatabases = len(dbFiltered)

	filename := e.service.GetBackupOptions().File.Path
	fullOutputPath := filepath.Join(e.service.GetBackupOptions().OutputDir, filename)
	logger.Debugf("Backup file: %s", fullOutputPath)

	// Capture GTID sebelum backup dimulai
	if err := e.service.CaptureAndSaveGTID(ctx, fullOutputPath); err != nil {
		logger.Warnf("GTID handling error: %v", err)
	}

	// Execute backup - gunakan mode dari BackupOptions (bisa 'all' atau 'combined')
	backupMode := e.service.GetBackupOptions().Mode
	backupInfo, execErr := e.service.ExecuteAndBuildBackup(ctx, types_backup.BackupExecutionConfig{
		DBList:       dbFiltered,
		OutputPath:   fullOutputPath,
		BackupType:   backupMode,
		TotalDBFound: totalDBFound,
		IsMultiDB:    true,
	})

	if execErr != nil {
		res.Errors = append(res.Errors, execErr.Error())
		res.FailedBackups = len(dbFiltered)
		for _, dbName := range dbFiltered {
			res.FailedDatabaseInfos = append(res.FailedDatabaseInfos, types_backup.FailedDatabaseInfo{
				DatabaseName: dbName,
				Error:        execErr.Error(),
			})
		}
		return res
	}

	// Success - all databases backed up in one file
	res.SuccessfulBackups = len(dbFiltered)

	// Export user grants:
	// - backup all: export semua user (pass nil)
	// - backup filter --mode=single-file: export hanya user dengan grants ke database yang dipilih (pass dbFiltered)
	var databasesToFilter []string
	if e.service.GetBackupOptions().Filter.IsFilterCommand {
		// Command filter: filter berdasarkan database yang dipilih
		databasesToFilter = dbFiltered
	}
	// Command all: nil (export semua user)

	actualUserGrantsPath := e.service.ExportUserGrantsIfNeeded(ctx, fullOutputPath, databasesToFilter)
	// Update metadata dengan actual path (atau "none" jika gagal)
	e.service.UpdateMetadataUserGrantsPath(fullOutputPath, actualUserGrantsPath)

	// Format display name dengan helper
	backupInfo.DatabaseName = e.formatCombinedBackupDisplayName(dbFiltered)

	res.BackupInfo = append(res.BackupInfo, backupInfo)
	return res
}

// formatCombinedBackupDisplayName memformat nama display untuk combined backup
func (e *CombinedExecutor) formatCombinedBackupDisplayName(databases []string) string {
	if len(databases) <= 10 {
		dbList := make([]string, len(databases))
		for i, db := range databases {
			dbList[i] = fmt.Sprintf("- %s", db)
		}
		return fmt.Sprintf("Combined backup (%d databases):\n%s", len(databases), strings.Join(dbList, "\n"))
	}
	return fmt.Sprintf("Combined backup (%d databases)", len(databases))
}
