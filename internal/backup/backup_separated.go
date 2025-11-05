package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"time"
)

// executeBackupSeparated melakukan backup setiap database dalam file terpisah
func (s *Service) executeBackupSeparated(ctx context.Context, dbFiltered []string) types.BackupResult {
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

		backupStartTime := time.Now()
		s.Log.Infof(fmt.Sprintf("[%d/%d] Backup database: %s", i+1, totalDatabases, dbName))

		s.BackupDBOptions.NamePattern = s.Config.Backup.Output.NamePattern

		// Generate filename untuk database ini
		filename, err := helper.GenerateBackupFilename(
			s.BackupDBOptions.NamePattern,
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
		writeResult, err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, fullOutputPath, Compression, CompressionType)

		if err != nil {
			// Hapus file backup yang gagal/kosong
			if _, statErr := os.Stat(fullOutputPath); statErr == nil {
				s.Log.Infof("Menghapus file backup yang gagal: %s", fullOutputPath)
				os.Remove(fullOutputPath)
			}

			errorMsg := fmt.Sprintf("gagal backup database %s: %v", dbName, err)
			if writeResult != nil && writeResult.StderrOutput != "" {
				errorMsg = fmt.Sprintf("%s\nDetail: %s", errorMsg, writeResult.StderrOutput)
			}

			s.Log.Error(errorMsg)
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
		backupDuration := time.Since(backupStartTime)
		fileInfo, err := os.Stat(fullOutputPath)
		var fileSize int64
		if err == nil {
			fileSize = fileInfo.Size()
		}

		// Generate metadata manifest dengan checksums jika enabled
		if s.Config.Backup.Verification.CompareChecksums && writeResult != nil {
			// Write metadata manifest (checksums sudah included di dalam JSON)
			metadata := s.CreateBackupMetadata(
				fullOutputPath,
				"separated",
				[]string{dbName},
				backupStartTime,
				time.Now(),
				writeResult.SHA256Hash,
				writeResult.MD5Hash,
				fileSize,
			)
			if err := s.WriteBackupMetadata(metadata); err != nil {
				s.Log.Warnf("Gagal menulis metadata file untuk %s: %v", dbName, err)
			}

			// Log checksum untuk reference
			s.Log.Debugf("Checksums untuk %s: SHA256=%s..., MD5=%s...",
				dbName, writeResult.SHA256Hash[:16], writeResult.MD5Hash[:16])

			// Verify checksum jika enabled
			if s.Config.Backup.Verification.VerifyAfterWrite {
				encryptionKey := ""
				if s.BackupDBOptions.Encryption.Enabled {
					encryptionKey = s.BackupDBOptions.Encryption.Key
				}

				compressionType := ""
				if s.BackupDBOptions.Compression.Enabled {
					compressionType = s.BackupDBOptions.Compression.Type
				}

				verifyResult, err := s.VerifyBackupChecksum(
					fullOutputPath,
					writeResult.SHA256Hash,
					writeResult.MD5Hash,
					encryptionKey,
					compressionType,
				)

				if err != nil {
					s.Log.Errorf("✗ Verifikasi checksum %s gagal: %v", dbName, err)
				} else if verifyResult.Success {
					s.Log.Debugf("✓ Verifikasi checksum %s BERHASIL", dbName)
				} else {
					s.Log.Errorf("✗ Verifikasi checksum %s GAGAL: %s", dbName, verifyResult.Error)
				}
			}
		}

		// Tentukan status berdasarkan stderr output
		backupStatus := "success"
		if stderrOutput != "" {
			backupStatus = "success_with_warnings"
			s.Log.Warningf(fmt.Sprintf("Database %s backup dengan warning: %s", dbName, stderrOutput))
		} else {
			s.Log.Info(fmt.Sprintf("✓ Database %s berhasil di-backup", dbName))
		}

		res.BackupInfo = append(res.BackupInfo, types.DatabaseBackupInfo{
			DatabaseName:  dbName,
			OutputFile:    fullOutputPath,
			FileSize:      fileSize,
			FileSizeHuman: global.FormatFileSize(fileSize),
			Duration:      global.FormatDuration(backupDuration),
			Status:        backupStatus,
			Warnings:      stderrOutput,
		})

		successCount++
	}

	// Clear current backup file setelah selesai
	s.ClearCurrentBackupFile()

	// Summary
	ui.PrintSubHeader("Ringkasan Backup Separated")
	ui.PrintInfo(fmt.Sprintf("Total database: %d", totalDatabases))
	ui.PrintSuccess(fmt.Sprintf("Berhasil: %d", successCount))
	if failedCount > 0 {
		ui.PrintError(fmt.Sprintf("Gagal: %d", failedCount))
	}

	return res
}
