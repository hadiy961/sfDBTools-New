package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"time"
)

// executeBackupSeparated melakukan backup setiap database dalam file terpisah
func (s *Service) ExecuteBackupSeparated(ctx context.Context, dbFiltered []string) types.BackupResult {
	var res types.BackupResult
	s.Log.Info("Melakukan backup database dalam mode separated")

	totalDatabases := len(dbFiltered)
	successCount := 0
	failedCount := 0

	// Dapatkan hostname untuk digunakan dalam filename
	dbHostname := s.BackupDBOptions.Profile.DBInfo.Host

	// Convert compression type
	compressionType := compress.CompressionNone
	if s.BackupDBOptions.Compression.Enabled {
		compressionType = compress.CompressionType(s.BackupDBOptions.Compression.Type)
	}

	s.Log.Infof(fmt.Sprintf("Memulai backup %d database secara terpisah...", totalDatabases))

	// Loop untuk setiap database
	for i, dbName := range dbFiltered {
		// Cek context untuk graceful shutdown
		select {
		case <-ctx.Done():
			s.Log.Warn("Proses backup dibatalkan oleh user")
			res.Errors = append(res.Errors, "Backup dibatalkan oleh user")
			return res
		default:
		}

		dbTimer := helper.NewTimer()
		s.Log.Infof(fmt.Sprintf("[%d/%d] Backup database: %s", i+1, totalDatabases, dbName))

		// Generate filename untuk database ini menggunakan fixed pattern
		filename, err := helper.GenerateBackupFilename(
			dbName,
			s.BackupDBOptions.Mode,
			dbHostname,
			compressionType,
			s.BackupDBOptions.Encryption.Enabled,
		)
		if err != nil {
			errorMsg := fmt.Sprintf("gagal generate filename untuk database %s: %v", dbName, err)
			s.Log.Error(errorMsg)
			res.FailedDatabaseInfos = append(res.FailedDatabaseInfos, types.FailedDatabaseInfo{
				DatabaseName: dbName,
				Error:        errorMsg,
			})
			failedCount++
			continue
		}

		fullOutputPath := filepath.Join(s.BackupDBOptions.OutputDir, filename)
		s.Log.Infof("Backup file: %s", fullOutputPath)

		// Set file backup yang sedang dibuat untuk graceful shutdown
		s.SetCurrentBackupFile(fullOutputPath)

		// Build mysqldump args untuk database tunggal
		mysqldumpArgs := s.buildMysqldumpArgs(s.Config.Backup.MysqlDumpArgs, nil, dbName, totalDatabases)

		// Eksekusi mysqldump
		Compression := s.BackupDBOptions.Compression.Enabled
		CompressionType := s.BackupDBOptions.Compression.Type
		// Capture start time for this database backup
		startTime := dbTimer.StartTime()

		writeResult, err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, fullOutputPath, Compression, CompressionType)

		if err != nil {
			// Hapus file backup yang gagal/kosong
			if fsops.FileExists(fullOutputPath) {
				s.Log.Infof("Menghapus file backup yang gagal: %s", fullOutputPath)
				os.Remove(fullOutputPath)
			}

			errorMsg := fmt.Sprintf("gagal backup database %s: %v", dbName, err)
			stderrDetail := ""
			if writeResult != nil && writeResult.StderrOutput != "" {
				errorMsg = fmt.Sprintf("%s\nDetail: %s", errorMsg, writeResult.StderrOutput)
				stderrDetail = writeResult.StderrOutput
			}

			s.Log.Error(errorMsg)

			// Log ke error log file terpisah (tidak tampilkan message berulang)
			if stderrDetail != "" {
				logFile := s.ErrorLog.LogWithOutput(map[string]interface{}{
					"database": dbName,
					"type":     "separated_backup",
					"file":     fullOutputPath,
				}, stderrDetail, err)
				if logFile != "" {
					// Log hanya sekali di awal, tidak di setiap database
					_ = logFile
				}
			}

			ui.PrintError(fmt.Sprintf("✗ Database %s gagal di-backup", dbName))
			res.FailedDatabaseInfos = append(res.FailedDatabaseInfos, types.FailedDatabaseInfo{
				DatabaseName: dbName,
				Error:        errorMsg,
			})
			failedCount++
			continue
		}

		stderrOutput := writeResult.StderrOutput

		// Backup berhasil
		backupDuration := dbTimer.Elapsed()
		endTime := time.Now()
		fileInfo, err := os.Stat(fullOutputPath)
		var fileSize int64
		if err == nil {
			fileSize = fileInfo.Size()
		}

		// Tentukan status berdasarkan stderr output
		backupStatus := "success"
		if stderrOutput != "" {
			backupStatus = "success_with_warnings"
			s.Log.Warningf(fmt.Sprintf("Database %s backup dengan warning: %s", dbName, stderrOutput))
		} else {
			s.Log.Info(fmt.Sprintf("✓ Database %s berhasil di-backup", dbName))
		}

		// Prepare manifest and optionally write it
		meta := types.BackupMetadata{
			BackupFile:      fullOutputPath,
			BackupType:      "separated",
			DatabaseNames:   []string{dbName},
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

		backupID := fmt.Sprintf("bk-%d", time.Now().UnixNano())
		var throughputMBs float64
		if backupDuration.Seconds() > 0 {
			throughputMBs = float64(fileSize) / (1024.0 * 1024.0) / backupDuration.Seconds()
		}

		res.BackupInfo = append(res.BackupInfo, types.DatabaseBackupInfo{
			DatabaseName:   dbName,
			OutputFile:     fullOutputPath,
			FileSize:       fileSize,
			FileSizeHuman:  global.FormatFileSize(fileSize),
			Duration:       global.FormatDuration(backupDuration),
			Status:         backupStatus,
			Warnings:       stderrOutput,
			BackupID:       backupID,
			StartTime:      startTime,
			EndTime:        endTime,
			ThroughputMBps: throughputMBs,
			ManifestFile:   manifestPath,
		})

		successCount++
	}

	// Clear current backup file setelah selesai
	s.ClearCurrentBackupFile()

	// Export user grants jika flag TIDAK exclude (ExcludeUser = false berarti export)
	if !s.BackupDBOptions.ExcludeUser && successCount > 0 {
		s.Log.Info("Export user grants ke file...")
		// Gunakan file backup pertama yang berhasil sebagai referensi untuk nama file
		if len(res.BackupInfo) > 0 {
			firstBackupFile := res.BackupInfo[0].OutputFile
			if userErr := s.ExportAndSaveUserGrants(ctx, firstBackupFile); userErr != nil {
				s.Log.Errorf("Gagal export user grants: %v", userErr)
				// Tidak fatal, lanjutkan
			}
		}
	}

	return res
}
