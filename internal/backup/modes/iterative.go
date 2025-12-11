// File : internal/backup/modes/iterative.go
// Deskripsi : Executor untuk mode backup yang bersifat iteratif (Single, Primary, Secondary, Separated)
//              Menggabungkan logika single.go dan separated.go untuk mengurangi duplikasi.
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-11
// Last Modified : 2025-12-11

package modes

import (
	"context"
	"errors"
	"path/filepath"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/backuphelper"
)

// IterativeExecutor menangani backup yang dilakukan per-database secara berurutan
// Digunakan untuk mode: single, primary, secondary, dan separated
type IterativeExecutor struct {
	service BackupService
	mode    string
}

// NewIterativeExecutor membuat instance baru IterativeExecutor
func NewIterativeExecutor(svc BackupService, mode string) *IterativeExecutor {
	return &IterativeExecutor{
		service: svc,
		mode:    mode,
	}
}

// Execute menjalankan backup secara iteratif
func (e *IterativeExecutor) Execute(ctx context.Context, dbList []string) types_backup.BackupResult {
	e.service.LogInfo("Melakukan backup database dalam mode " + e.mode)

	isSeparated := e.mode == "separated" || e.mode == "separate"
	
	// Khusus mode separated: Capture GTID global di awal untuk konsistensi (snapshot point)
	if isSeparated && len(dbList) > 0 {
		// Gunakan nama database pertama untuk generate path reference file GTID
		firstPath, _ := e.service.GenerateFullBackupPath(dbList[0], e.service.GetBackupOptions().Mode)
		if err := e.service.CaptureAndSaveGTID(ctx, firstPath); err != nil {
			e.service.LogWarn("GTID handling error: " + err.Error())
		}
	}

	// Siapkan fungsi penentu path output
	outputPathFunc := e.createOutputPathFunc(dbList)

	// Jalankan backup loop
	loopResult := e.service.ExecuteBackupLoop(ctx, dbList, types_backup.BackupLoopConfig{
		Mode:       e.mode,
		TotalDBs:   len(dbList),
		BackupType: e.mode,
	}, outputPathFunc)

	// Convert ke BackupResult standar
	res := e.service.ToBackupResult(loopResult)

	// Export user grants jika ada setidaknya satu backup yang berhasil
	// File user grants akan diletakkan bersebelahan dengan file backup pertama yang sukses
	if len(loopResult.BackupInfos) > 0 {
		e.service.ExportUserGrantsIfNeeded(ctx, loopResult.BackupInfos[0].OutputFile)
	}

	// Update statistik akhir
	res.TotalDatabases = len(dbList)
	res.SuccessfulBackups = loopResult.Success
	res.FailedBackups = loopResult.Failed

	// Khusus Single Mode variant: Jika semua gagal, pastikan ada error eksplisit
	if backuphelper.IsSingleModeVariant(e.mode) && loopResult.Success == 0 && len(res.Errors) == 0 && len(res.FailedDatabaseInfos) > 0 {
		res.Errors = append(res.Errors, errors.New("backup gagal untuk semua database").Error())
	}

	return res
}

// createOutputPathFunc membuat fungsi closure untuk menentukan path output setiap database
func (e *IterativeExecutor) createOutputPathFunc(dbList []string) func(string) (string, error) {
	primaryFilename := e.service.GetBackupOptions().File.Filename
	
	return func(dbName string) (string, error) {
		// Logika khusus untuk Single Mode Variant (Single, Primary, Secondary):
		// Database pertama (index 0) bisa menggunakan custom filename jika diset user.
		// Companion databases (dmart, temp, archive) akan tetap digenerate namanya.
		if backuphelper.IsSingleModeVariant(e.mode) && len(dbList) > 0 && dbList[0] == dbName && primaryFilename != "" {
			return filepath.Join(e.service.GetBackupOptions().OutputDir, primaryFilename), nil
		}
		
		// Default: generate full path berdasarkan pattern standar
		return e.service.GenerateFullBackupPath(dbName, e.service.GetBackupOptions().Mode)
	}
}
