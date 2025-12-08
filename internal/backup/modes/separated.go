// File : internal/backup/modes/separated.go
// Deskripsi : Mode backup separated - setiap database dalam file terpisah
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package modes

import (
	"context"
	"sfDBTools/internal/types/types_backup"
)

// SeparatedExecutor menangani backup separated mode
type SeparatedExecutor struct {
	service BackupService
}

// NewSeparatedExecutor membuat instance baru SeparatedExecutor
func NewSeparatedExecutor(svc BackupService) *SeparatedExecutor {
	return &SeparatedExecutor{service: svc}
}

// Execute melakukan backup setiap database dalam file terpisah
func (e *SeparatedExecutor) Execute(ctx context.Context, dbFiltered []string) types_backup.BackupResult {
	e.service.LogInfo("Melakukan backup database dalam mode separated")

	// Jalankan backup loop
	loopResult := e.service.ExecuteBackupLoop(ctx, dbFiltered, types_backup.BackupLoopConfig{
		Mode:       "separated",
		TotalDBs:   len(dbFiltered),
		BackupType: "separated",
	}, func(dbName string) (string, error) {
		return e.service.GenerateFullBackupPath(dbName, e.service.GetBackupOptions().Mode)
	})

	// Convert ke BackupResult
	res := e.service.ToBackupResult(loopResult)
	res.TotalDatabases = len(dbFiltered)
	res.SuccessfulBackups = loopResult.Success
	res.FailedBackups = loopResult.Failed

	// Export user grants jika ada backup yang berhasil
	if len(loopResult.BackupInfos) > 0 {
		e.service.ExportUserGrantsIfNeeded(ctx, loopResult.BackupInfos[0].OutputFile)
	}

	return res
}
