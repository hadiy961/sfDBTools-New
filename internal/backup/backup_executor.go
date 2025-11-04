package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/ui"
	"strings"
	"time"
)

// ExecuteBackup melakukan proses backup database
func (s *Service) ExecuteBackup(ctx context.Context, sourceClient *database.Client, dbFiltered []string, backupMode string) (*types.BackupResult, error) {

	// Simpan client ke service untuk digunakan di fungsi lain
	s.Client = sourceClient

	// 1. Setup konfigurasi backup
	err := s.SetupBackupExecution()
	if err != nil {
		return nil, fmt.Errorf("gagal setup backup execution: %w", err)
	}

	// 2. Eksekusi backup sesuai mode
	startTime := time.Now()
	var result types.BackupResult

	ui.PrintSubHeader("Memulai Proses Backup")

	if backupMode == "separate" {
		s.Log.Error("Fitur belum di implementasikan")
		return nil, fmt.Errorf("fitur backup terpisah belum di implementasikan")
	} else {
		result = s.executeBackupCombined(ctx, dbFiltered)
	}

	result.TotalTimeTaken = time.Since(startTime)

	// 3. Cek apakah ada error dalam result
	if len(result.Errors) > 0 || len(result.FailedDatabaseInfos) > 0 {
		// Jika ada error, kembalikan sebagai error
		return &types.BackupResult{
			TotalDatabases:      len(dbFiltered),
			SuccessfulBackups:   0,
			FailedBackups:       len(dbFiltered),
			FailedDatabases:     map[string]string{},
			BackupInfo:          result.BackupInfo,
			FailedDatabaseInfos: result.FailedDatabaseInfos,
			Errors:              result.Errors,
			TotalTimeTaken:      result.TotalTimeTaken,
		}, fmt.Errorf("backup gagal: %s", result.Errors[0])
	}

	// 4. Kembalikan hasil backup sukses
	return &types.BackupResult{
		TotalDatabases:    len(dbFiltered),
		SuccessfulBackups: len(dbFiltered),
		FailedBackups:     0,
		FailedDatabases:   map[string]string{},
		BackupInfo:        result.BackupInfo,
		TotalTimeTaken:    result.TotalTimeTaken,
	}, nil
}

// executeBackupCombined melakukan backup semua database dalam satu file
func (s *Service) executeBackupCombined(ctx context.Context, dbFiltered []string) types.BackupResult {
	backupStartTime := time.Now()
	var res types.BackupResult
	s.Log.Info("Melakukan backup database dalam mode combined")

	// Dapatkan total database dari server untuk logika --all-databases
	allDatabases, err := s.Client.GetDatabaseList(ctx)
	totalDBFound := len(allDatabases)
	if err != nil {
		s.Log.Warnf("Gagal mendapatkan total database: %v, menggunakan fallback", err)
		totalDBFound = len(dbFiltered) // Fallback ke jumlah filtered
	}

	mysqldumpArgs := s.buildMysqldumpArgs(s.Config.Backup.MysqlDumpArgs, dbFiltered, "", totalDBFound)

	filename := s.BackupDBOptions.File.Path
	fullOutputPath := filepath.Join(s.BackupDBOptions.OutputDir, filename)
	s.Log.Info("Generated backup filename: " + fullOutputPath)

	Compression := s.BackupDBOptions.Compression.Enabled
	CompressionType := s.BackupDBOptions.Compression.Type
	stderrOutput, err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, fullOutputPath, Compression, CompressionType)
	if err != nil {
		// Hapus file backup yang gagal/kosong
		if _, statErr := os.Stat(fullOutputPath); statErr == nil {
			s.Log.Infof("Menghapus file backup yang gagal: %s", fullOutputPath)
			os.Remove(fullOutputPath)
		}

		// Tampilkan error detail dari mysqldump
		ui.PrintHeader("ERROR : Mysqldump Gagal dijalankan")

		if stderrOutput != "" {
			ui.PrintSubHeader("Detail Error dari mysqldump:")
			ui.PrintError(stderrOutput)
		}

		errorMsg := fmt.Errorf("gagal menjalankan mysqldump: %w", err)
		res.Errors = append(res.Errors, errorMsg.Error())
		for _, dbName := range dbFiltered {
			res.FailedDatabaseInfos = append(res.FailedDatabaseInfos, types.FailedDatabaseInfo{DatabaseName: dbName, Error: errorMsg.Error()})
		}
		return res
	} // Tentukan status berdasarkan stderr output (hanya untuk NON-FATAL warnings)
	// Jika sampai sini berarti mysqldump berhasil, tapi mungkin ada warning
	backupStatus := "success"
	if stderrOutput != "" {
		backupStatus = "success_with_warnings"
		ui.PrintWarning("Backup selesai dengan warning (data ter-backup tapi ada pesan):")
		ui.PrintWarning(stderrOutput)
	}

	backupDuration := time.Since(backupStartTime)
	fileInfo, err := os.Stat(fullOutputPath)
	var fileSize int64
	if err == nil {
		fileSize = fileInfo.Size()
	}

	// Untuk combined mode, buat satu entry dengan info semua database
	dbListStr := fmt.Sprintf("Combined backup (%d databases)", len(dbFiltered))
	if len(dbFiltered) <= 10 {
		// Tampilkan nama database jika tidak terlalu banyak
		dbListStr = fmt.Sprintf("%d databases: %s", len(dbFiltered), strings.Join(dbFiltered, ", "))
	}

	res.BackupInfo = append(res.BackupInfo, types.DatabaseBackupInfo{
		DatabaseName:  dbListStr,
		OutputFile:    fullOutputPath,
		FileSize:      fileSize,
		FileSizeHuman: global.FormatFileSize(fileSize),
		Duration:      global.FormatDuration(backupDuration),
		Status:        backupStatus,
		Warnings:      stderrOutput,
	})

	return res
}
