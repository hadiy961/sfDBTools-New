// File : internal/backup/modes/iterative.go
// Deskripsi : Executor untuk mode backup yang bersifat iteratif (Single, Primary, Secondary, Separated)
//              Menggabungkan logika single.go dan separated.go untuk mengurangi duplikasi.
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-11
// Last Modified : 2026-01-02

package modes

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/app/backup/metadata"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/consts"
	"strings"
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
	e.service.GetLog().Info("Melakukan backup database dalam mode " + e.mode)

	// Untuk mode separated/multi-file: TIDAK capture GTID karena setiap database dibackup terpisah
	// dan tidak ada konsep snapshot point global yang relevan
	// GTID capture hanya dilakukan untuk mode combined dan single-mode variants

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

	// Export user grants dan generate metadata untuk primary/secondary:
	// - Mode separated/single: sudah di-handle per database di executeBackupLoop
	// - Mode primary/secondary: export satu file dan satu metadata untuk semua database
	if len(loopResult.BackupInfos) > 0 && (e.mode == consts.ModePrimary || e.mode == consts.ModeSecondary) {
		// Untuk primary/secondary, export user grants yang punya akses ke database dalam list
		actualUserGrantsPath := e.service.ExportUserGrantsIfNeeded(ctx, loopResult.BackupInfos[0].OutputFile, dbList)
		// Update metadata dengan actual path (atau "none" jika gagal)
		e.service.UpdateMetadataUserGrantsPath(loopResult.BackupInfos[0].OutputFile, actualUserGrantsPath)

		// Generate satu metadata untuk semua database yang berhasil di-backup
		e.generateCombinedMetadata(loopResult, dbList)

		// Aggregate backup infos menjadi satu entry untuk display
		res.BackupInfo = e.aggregateBackupInfos(loopResult.BackupInfos)
	}

	// Update statistik akhir
	res.TotalDatabases = len(dbList)
	res.SuccessfulBackups = loopResult.Success
	res.FailedBackups = loopResult.Failed

	// Khusus Single Mode variant: Jika semua gagal, pastikan ada error eksplisit
	if IsSingleModeVariant(e.mode) && loopResult.Success == 0 && len(res.Errors) == 0 && len(res.FailedDatabaseInfos) > 0 {
		res.Errors = append(res.Errors, errors.New("backup gagal untuk semua database").Error())
	}

	return res
}

// createOutputPathFunc membuat fungsi closure untuk menentukan path output setiap database
func (e *IterativeExecutor) createOutputPathFunc(dbList []string) func(string) (string, error) {
	opts := e.service.GetOptions()
	primaryFilename := opts.File.Filename
	primaryDBName := ""
	if len(dbList) > 0 {
		primaryDBName = dbList[0]
	}

	// Precompute final filename for primary DB if custom filename is set.
	primaryFinalFilename := ""
	if primaryDBName != "" && primaryFilename != "" {
		defaultPath, derr := e.service.GenerateFullBackupPath(primaryDBName, opts.Mode)
		if derr == nil {
			defaultName := filepath.Base(defaultPath)
			primaryFinalFilename = applyCustomBaseFilename(defaultName, primaryFilename)
		}
	}

	return func(dbName string) (string, error) {
		// Logika khusus untuk Single Mode Variant (Single, Primary, Secondary):
		// Database pertama (index 0) bisa menggunakan custom filename jika diset user.
		// Companion databases (dmart, temp, archive) akan tetap digenerate namanya.
		if IsSingleModeVariant(e.mode) && primaryDBName != "" && primaryDBName == dbName && strings.TrimSpace(primaryFilename) != "" {
			// Auto-append extension chain based on default generated filename.
			if strings.TrimSpace(primaryFinalFilename) != "" {
				return filepath.Join(opts.OutputDir, primaryFinalFilename), nil
			}
			defaultPath, derr := e.service.GenerateFullBackupPath(dbName, opts.Mode)
			if derr != nil {
				return "", derr
			}
			defaultName := filepath.Base(defaultPath)
			finalName := applyCustomBaseFilename(defaultName, primaryFilename)
			return filepath.Join(opts.OutputDir, finalName), nil
		}

		// Khusus PRIMARY/SECONDARY: companion file mengikuti pola filename utama.
		// Contoh:
		// - custom filename: tes_secondary_123_123
		// - companion db: <primary>_dmart -> tes_secondary_123_123_dmart.sql.gz.enc
		if (e.mode == consts.ModePrimary || e.mode == consts.ModeSecondary) &&
			strings.TrimSpace(primaryFinalFilename) != "" &&
			primaryDBName != "" && dbName != primaryDBName {
			if suffix, ok := getCompanionSuffix(primaryDBName, dbName); ok {
				companionFilename := insertSuffixBeforeSQLExt(primaryFinalFilename, suffix)
				return filepath.Join(opts.OutputDir, companionFilename), nil
			}
		}

		// Default: generate full path berdasarkan pattern standar
		return e.service.GenerateFullBackupPath(dbName, opts.Mode)
	}
}

// aggregateBackupInfos menggabungkan multiple backup infos menjadi satu entry
// Digunakan untuk primary/secondary yang backup multiple databases tapi display sebagai satu
func (e *IterativeExecutor) aggregateBackupInfos(backupInfos []types_backup.DatabaseBackupInfo) []types_backup.DatabaseBackupInfo {
	if len(backupInfos) == 0 {
		return backupInfos
	}

	// Ambil info pertama sebagai base
	aggregated := backupInfos[0]

	// Update database name untuk menunjukkan multiple databases
	aggregated.DatabaseName = fmt.Sprintf("%s + %d companion databases", backupInfos[0].DatabaseName, len(backupInfos)-1)

	// Sum up file sizes dari semua backup
	totalSize := int64(0)
	for _, info := range backupInfos {
		totalSize += info.FileSize
	}
	aggregated.FileSize = totalSize
	aggregated.FileSizeHuman = fmt.Sprintf("%.2f MB", float64(totalSize)/(1024*1024))

	// Metadata file adalah dari database pertama (combined metadata)
	aggregated.ManifestFile = backupInfos[0].OutputFile + consts.ExtMetaJSON

	return []types_backup.DatabaseBackupInfo{aggregated}
}

// generateCombinedMetadata membuat satu metadata file untuk semua database yang berhasil di-backup
// Digunakan untuk mode primary/secondary yang backup multiple databases (main + companions)
func (e *IterativeExecutor) generateCombinedMetadata(loopResult types_backup.BackupLoopResult, dbList []string) {
	// Tidak generate metadata jika tidak ada backup yang berhasil
	if len(loopResult.BackupInfos) == 0 {
		return
	}

	e.service.GetLog().Debug(fmt.Sprintf("Generating combined metadata untuk %d databases", len(dbList)))

	// Untuk primary/secondary:
	// 1. Update metadata pertama dengan DatabaseNames dan DatabaseDetails (info lengkap per database)
	// 2. Hapus semua metadata individual untuk companion databases

	// Update metadata pertama dengan full database list dan details
	primaryBackupFile := loopResult.BackupInfos[0].OutputFile
	if err := metadata.UpdateMetadataWithDatabaseDetails(primaryBackupFile, dbList, loopResult.BackupInfos, e.service.GetLog()); err != nil {
		e.service.GetLog().Warn("Gagal update combined metadata: " + err.Error())
	}

	// Hapus metadata individual untuk companion databases (index 1+)
	for i, info := range loopResult.BackupInfos {
		if i == 0 {
			continue
		}
		// Companion databases: hapus metadata individual
		metadataPath := info.OutputFile + consts.ExtMetaJSON
		e.service.GetLog().Debug("Menghapus metadata companion: " + metadataPath)
		os.Remove(metadataPath)
	}
}
