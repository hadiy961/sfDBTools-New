package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"
	"time"
)

// executeBackupCombined melakukan backup semua database dalam satu file
func (s *Service) ExecuteBackupCombined(ctx context.Context, dbFiltered []string) types.BackupResult {
	timer := helper.NewTimer()
	var res types.BackupResult
	s.Log.Info("Melakukan backup database dalam mode combined")

	// Capture GTID SEBELUM backup dimulai (saat koneksi masih aktif)
	var gtidInfo *database.GTIDInfo
	if s.BackupDBOptions.CaptureGTID {
		s.Log.Info("Mengambil informasi GTID sebelum backup...")
		var gtidErr error
		gtidInfo, gtidErr = s.Client.GetFullGTIDInfo(ctx)
		if gtidErr != nil {
			s.Log.Warnf("Gagal mendapatkan GTID: %v", gtidErr)
			// Tidak fatal, lanjutkan backup
		} else {
			s.Log.Infof("GTID berhasil diambil: File=%s, Pos=%d", gtidInfo.MasterLogFile, gtidInfo.MasterLogPos)
		}
	}

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
	// Capture start time from timer
	startTime := timer.StartTime()

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
	endTime := time.Now()

	fileInfo, err := os.Stat(fullOutputPath)
	var fileSize int64
	if err == nil {
		fileSize = fileInfo.Size()
	}

	// Simpan GTID ke file jika sudah diambil sebelumnya
	if gtidInfo != nil {
		s.Log.Info("Menyimpan informasi GTID ke file...")
		if gtidErr := s.SaveGTIDToFile(gtidInfo, fullOutputPath); gtidErr != nil {
			s.Log.Errorf("Gagal menyimpan GTID ke file: %v", gtidErr)
			// Tidak fatal, lanjutkan
		}
	}

	// Export user grants jika flag TIDAK exclude (ExcludeUser = false berarti export)
	if !s.BackupDBOptions.ExcludeUser {
		s.Log.Info("Export user grants ke file...")
		if userErr := s.ExportAndSaveUserGrants(ctx, fullOutputPath); userErr != nil {
			s.Log.Errorf("Gagal export user grants: %v", userErr)
			// Tidak fatal, lanjutkan
		}
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

	// Compute throughput (MB/s)
	var throughputMBs float64
	if backupDuration.Seconds() > 0 {
		throughputMBs = float64(fileSize) / (1024.0 * 1024.0) / backupDuration.Seconds()
	}

	// Generate simple backup ID (time-based)
	backupID := fmt.Sprintf("bk-%d", time.Now().UnixNano())

	// Build manifest metadata (without checksum as requested)
	meta := types.BackupMetadata{
		BackupFile:      fullOutputPath,
		BackupType:      "combined",
		DatabaseNames:   dbFiltered,
		Hostname:        s.BackupDBOptions.Profile.DBInfo.HostName,
		BackupStartTime: startTime,
		BackupEndTime:   endTime,
		BackupDuration:  global.FormatDuration(backupDuration),
		FileSize:        fileSize,
		FileSizeHuman:   global.FormatFileSize(fileSize),
		Compressed:      Compression,
		CompressionType: CompressionType,
		Encrypted:       s.BackupDBOptions.Encryption.Enabled,
		BackupStatus:    backupStatus,
		Warnings:        []string{},
		GeneratedBy:     "sfDBTools",
		GeneratedAt:     time.Now(),
	}
	if stderrOutput != "" {
		meta.Warnings = append(meta.Warnings, stderrOutput)
	}

	// Write manifest to disk atomically only if configured to save backup info.
	// Support both `create_backup_info` and `save_backup_info` config keys.
	shouldSaveMeta := s.Config.Backup.Output.CreateBackupInfo || s.Config.Backup.Output.SaveBackupInfo

	manifestPath := ""
	if shouldSaveMeta {
		manifestPath = fullOutputPath + ".meta.json"
		tmpManifest := manifestPath + ".tmp"
		if manifestBytes, mErr := json.MarshalIndent(meta, "", "  "); mErr == nil {
			if wErr := os.WriteFile(tmpManifest, manifestBytes, 0644); wErr != nil {
				s.Log.Errorf("Gagal menulis file manifest sementara: %v", wErr)
				manifestPath = ""
			} else if rErr := os.Rename(tmpManifest, manifestPath); rErr != nil {
				s.Log.Errorf("Gagal merename manifest sementara ke target: %v", rErr)
				manifestPath = ""
			}
		} else {
			s.Log.Errorf("Gagal membuat manifest JSON: %v", mErr)
			manifestPath = ""
		}
	}

	// Append backup info; manifestPath will be empty string if not saved or on error
	res.BackupInfo = append(res.BackupInfo, types.DatabaseBackupInfo{
		DatabaseName:  dbListStr,
		OutputFile:    fullOutputPath,
		FileSize:      fileSize,
		FileSizeHuman: global.FormatFileSize(fileSize),
		Duration:      global.FormatDuration(backupDuration),
		Status:        backupStatus,
		Warnings:      stderrOutput,
		BackupID:      backupID,
		StartTime:     startTime,
		EndTime:       endTime,

		ThroughputMBps: throughputMBs,
		ManifestFile:   manifestPath,
	})

	return res
}
