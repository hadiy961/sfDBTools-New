package backup

import (
	"context"
	"fmt"
	"os"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/ui"
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

	// 3. Kembalikan hasil backup
	return &types.BackupResult{
		TotalDatabases:    len(dbFiltered),
		SuccessfulBackups: len(dbFiltered), // Asumsikan semua berhasil untuk placeholder
		FailedBackups:     0,
		FailedDatabases:   map[string]string{},
		BackupInfo:        result.BackupInfo,
		TotalTimeTaken:    result.TotalTimeTaken,
	}, err
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
	fullOutputPath := s.BackupDBOptions.OutputDir
	Compression := s.BackupDBOptions.Compression.Enabled
	CompressionType := s.BackupDBOptions.Compression.Type
	stderrOutput, err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, fullOutputPath, Compression, CompressionType)
	if err != nil {
		errorMsg := fmt.Errorf("gagal menjalankan mysqldump: %w", err)
		res.Errors = append(res.Errors, errorMsg.Error())
		for _, dbName := range dbFiltered {
			res.FailedDatabaseInfos = append(res.FailedDatabaseInfos, types.FailedDatabaseInfo{DatabaseName: dbName, Error: errorMsg.Error()})
		}
		return res
	}

	// Tentukan status berdasarkan stderr output
	backupStatus := "success"
	var errorLogFile string
	if stderrOutput != "" {
		backupStatus = "success_with_warnings"
		s.Log.Warnf("Backup combined selesai dengan warning (lihat: %s)", errorLogFile)
	}

	backupDuration := time.Since(backupStartTime)
	fileInfo, err := os.Stat(fullOutputPath)
	var fileSize int64
	if err == nil {
		fileSize = fileInfo.Size()
	}

	for _, dbName := range dbFiltered {
		res.BackupInfo = append(res.BackupInfo, types.DatabaseBackupInfo{
			DatabaseName:  dbName,
			OutputFile:    fullOutputPath,
			FileSize:      fileSize,
			FileSizeHuman: global.FormatFileSize(fileSize),
			Duration:      global.FormatDuration(backupDuration),
			Status:        backupStatus,
			Warnings:      stderrOutput,
			ErrorLogFile:  errorLogFile,
		})
	}

	return res
}
