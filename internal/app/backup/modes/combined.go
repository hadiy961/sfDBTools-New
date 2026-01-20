// File : internal/backup/modes/combined.go
// Deskripsi : Mode backup combined - semua database dalam satu file
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 20 Januari 2026

package modes

import (
	"context"
	"fmt"
	"path/filepath"
	backuppath "sfdbtools/internal/app/backup/helpers/path"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/shared/consts"
	"strings"
	"time"
)

// CombinedExecutor menangani backup combined mode
type CombinedExecutor struct {
	service BackupService
	state   BackupStateAccessor
}

// NewCombinedExecutor membuat instance baru CombinedExecutor
func NewCombinedExecutor(svc BackupService, state BackupStateAccessor) *CombinedExecutor {
	return &CombinedExecutor{
		service: svc,
		state:   state,
	}
}

// Execute melakukan backup semua database dalam satu file
func (e *CombinedExecutor) Execute(ctx context.Context, dbFiltered []string) types_backup.BackupResult {
	var res types_backup.BackupResult
	e.service.GetLog().Info("Melakukan backup database dalam mode combined")

	totalDBFound := len(dbFiltered)

	// Informasi database yang akan di-backup (penting untuk mode background/quiet)
	e.service.GetLog().Infof("Database yang akan di-backup (combined, total=%d)", totalDBFound)
	if totalDBFound <= consts.MaxDisplayDatabases {
		e.service.GetLog().Infof("Daftar database: %s", strings.Join(dbFiltered, ", "))
	} else {
		e.service.GetLog().Infof("Daftar database (first %d): %s", consts.MaxDisplayDatabases, strings.Join(dbFiltered[:consts.MaxDisplayDatabases], ", "))
		// e.service.GetLog().Debugf("Daftar database lengkap: %v", dbFiltered)
	}

	// Initialize result statistics
	res.TotalDatabases = len(dbFiltered)

	opts := e.service.GetOptions()
	filename := opts.File.Path
	// Mode all/combined bisa override lewat --filename (base name tanpa ekstensi).
	if (opts.Mode == consts.ModeAll || opts.Mode == consts.ModeCombined) && opts.File.Filename != "" {
		filename = backuppath.ApplyCustomBaseFilename(filename, opts.File.Filename)
	}
	fullOutputPath := filepath.Join(opts.OutputDir, filename)
	e.service.GetLog().Debug("Backup file: " + fullOutputPath)

	// Capture GTID sebelum backup dimulai
	if err := e.service.CaptureAndSaveGTID(ctx, e.state, fullOutputPath); err != nil {
		e.service.GetLog().Warn("GTID handling error: " + err.Error())
	}

	// Execute backup - gunakan mode dari BackupOptions (bisa 'all' atau 'combined')
	backupMode := opts.Mode
	start := time.Now()
	e.service.GetLog().Infof("Memulai dump combined untuk %d database", totalDBFound)
	backupInfo, execErr := e.service.ExecuteAndBuildBackup(ctx, e.state, types_backup.BackupExecutionConfig{
		DBList:       dbFiltered,
		OutputPath:   fullOutputPath,
		BackupType:   backupMode,
		TotalDBFound: totalDBFound,
		IsMultiDB:    true,
	})
	if execErr != nil {
		// Error untuk combined/all dilaporkan oleh layer command agar tidak duplikasi/spam.
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
	e.service.GetLog().Infof("Selesai dump combined (%s)", time.Since(start).Round(time.Millisecond))

	// Success - all databases backed up in one file
	res.SuccessfulBackups = len(dbFiltered)

	// Export user grants:
	// - backup all: export semua user (pass nil)
	// - backup filter --mode=single-file: export hanya user dengan grants ke database yang dipilih (pass dbFiltered)
	var databasesToFilter []string
	if e.service.GetOptions().Filter.IsFilterCommand {
		// Command filter: filter berdasarkan database yang dipilih
		databasesToFilter = dbFiltered
	}
	// Command all: nil (export semua user)

	actualUserGrantsPath := e.service.ExportUserGrantsIfNeeded(ctx, fullOutputPath, databasesToFilter)
	// Update metadata dengan actual path (atau "none" jika gagal)
	permissions := e.service.GetConfig().Backup.Output.MetadataPermissions
	e.service.UpdateMetadataUserGrantsPath(fullOutputPath, actualUserGrantsPath, permissions)

	// Format display name dengan helper
	backupInfo.DatabaseName = e.formatCombinedBackupDisplayName(dbFiltered)

	res.BackupInfo = append(res.BackupInfo, backupInfo)
	return res
}

// formatCombinedBackupDisplayName memformat nama display untuk combined backup
func (e *CombinedExecutor) formatCombinedBackupDisplayName(databases []string) string {
	if len(databases) <= consts.MaxDisplayDatabases {
		dbList := make([]string, len(databases))
		for i, db := range databases {
			dbList[i] = fmt.Sprintf("- %s", db)
		}
		return fmt.Sprintf("Combined backup (%d databases):\n%s", len(databases), strings.Join(dbList, "\n"))
	}
	return fmt.Sprintf("Combined backup (%d databases)", len(databases))
}
