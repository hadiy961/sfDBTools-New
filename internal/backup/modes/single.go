// File : internal/backup/modes/single.go
// Deskripsi : Mode backup single - satu database utama dengan optional companion
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package modes

import (
	"context"
	"errors"
	"path/filepath"
	"sfDBTools/internal/types/types_backup"
)

// SingleExecutor menangani backup single mode
type SingleExecutor struct {
	service BackupService
}

// NewSingleExecutor membuat instance baru SingleExecutor
func NewSingleExecutor(svc BackupService) *SingleExecutor {
	return &SingleExecutor{service: svc}
}

// Execute melakukan backup untuk satu database dan optional companion databases
// Companion databases berasal dari flag include-dmart/temp/archive dan hanya di-backup jika ada
func (e *SingleExecutor) Execute(ctx context.Context, dbList []string) types_backup.BackupResult {
	e.service.LogInfo("Melakukan backup database dalam mode single")

	primaryFilename := e.service.GetBackupOptions().File.Filename

	// Helper function untuk generate output path dengan special handling untuk primary db
	outputPathFunc := func(dbName string) (string, error) {
		// Gunakan custom filename untuk database utama (index 0)
		if dbList[0] == dbName && primaryFilename != "" {
			return filepath.Join(e.service.GetBackupOptions().OutputDir, primaryFilename), nil
		}
		return e.service.GenerateFullBackupPath(dbName, e.service.GetBackupOptions().Mode)
	}

	// Jalankan backup loop
	loopResult := e.service.ExecuteBackupLoop(ctx, dbList, types_backup.BackupLoopConfig{
		Mode:       "single",
		TotalDBs:   len(dbList),
		BackupType: "single",
	}, outputPathFunc)

	// Convert ke BackupResult
	res := e.service.ToBackupResult(loopResult)

	// Export user grants jika ada backup yang berhasil
	if len(loopResult.BackupInfos) > 0 {
		e.service.ExportUserGrantsIfNeeded(ctx, loopResult.BackupInfos[0].OutputFile)
	}

	res.TotalDatabases = len(dbList)
	res.SuccessfulBackups = loopResult.Success
	res.FailedBackups = loopResult.Failed

	if loopResult.Success == 0 && len(res.Errors) == 0 && len(res.FailedDatabaseInfos) > 0 {
		res.Errors = append(res.Errors, errors.New("backup gagal untuk semua database").Error())
	}

	return res
}
