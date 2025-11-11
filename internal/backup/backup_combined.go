package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"
)

// executeBackupCombined melakukan backup semua database dalam satu file
func (s *Service) ExecuteBackupCombined(ctx context.Context, dbFiltered []string) types.BackupResult {
	timer := helper.NewTimer()
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

	// Set file backup yang sedang dibuat untuk graceful shutdown
	s.SetCurrentBackupFile(fullOutputPath)
	defer s.ClearCurrentBackupFile()

	Compression := s.BackupDBOptions.Compression.Enabled
	CompressionType := s.BackupDBOptions.Compression.Type
	writeResult, err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, fullOutputPath, Compression, CompressionType)
	if err != nil {
		// Hapus file backup yang gagal/kosong
		if fsops.FileExists(fullOutputPath) {
			s.Log.Infof("Menghapus file backup yang gagal: %s", fullOutputPath)
			os.Remove(fullOutputPath)
		}

		// Tampilkan error detail dari mysqldump
		ui.PrintHeader("ERROR : Mysqldump Gagal dijalankan")

		stderrDetail := ""
		if writeResult != nil && writeResult.StderrOutput != "" {
			ui.PrintSubHeader("Detail Error dari mysqldump:")
			ui.PrintError(writeResult.StderrOutput)
			stderrDetail = writeResult.StderrOutput
		}

		errorMsg := fmt.Errorf("gagal menjalankan mysqldump: %w", err)

		// Log ke error log file terpisah
		logFile := s.ErrorLog.LogWithOutput(map[string]interface{}{
			"type": "combined_backup",
			"file": fullOutputPath,
		}, stderrDetail, err)
		_ = logFile

		res.Errors = append(res.Errors, errorMsg.Error())
		for _, dbName := range dbFiltered {
			res.FailedDatabaseInfos = append(res.FailedDatabaseInfos, types.FailedDatabaseInfo{DatabaseName: dbName, Error: errorMsg.Error()})
		}
		return res
	}

	stderrOutput := writeResult.StderrOutput

	// Tentukan status berdasarkan stderr output (hanya untuk NON-FATAL warnings)
	// Jika sampai sini berarti mysqldump berhasil, tapi mungkin ada warning
	backupStatus := "success"
	if stderrOutput != "" {
		backupStatus = "success_with_warnings"
		ui.PrintWarning("Backup selesai dengan warning (data ter-backup tapi ada pesan):")
		ui.PrintWarning(stderrOutput)
	}

	backupDuration := timer.Elapsed()
	fileInfo, err := os.Stat(fullOutputPath)
	var fileSize int64
	if err == nil {
		fileSize = fileInfo.Size()
	}

	// Untuk combined mode, buat satu entry dengan info semua database
	var dbListStr string
	if len(dbFiltered) <= 10 {
		// Tampilkan nama database dalam format bullet list jika tidak terlalu banyak
		dbList := make([]string, len(dbFiltered))
		for i, db := range dbFiltered {
			dbList[i] = fmt.Sprintf("- %s", db)
		}
		dbListStr = fmt.Sprintf("Combined backup (%d databases):\n%s", len(dbFiltered), strings.Join(dbList, "\n"))
	} else {
		dbListStr = fmt.Sprintf("Combined backup (%d databases)", len(dbFiltered))
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
