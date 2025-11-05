package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/ui"
	"strings"
	"time"
)

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

	// Set file backup yang sedang dibuat untuk graceful shutdown
	s.SetCurrentBackupFile(fullOutputPath)
	defer s.ClearCurrentBackupFile()

	Compression := s.BackupDBOptions.Compression.Enabled
	CompressionType := s.BackupDBOptions.Compression.Type
	writeResult, err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, fullOutputPath, Compression, CompressionType)
	if err != nil {
		// Hapus file backup yang gagal/kosong
		if _, statErr := os.Stat(fullOutputPath); statErr == nil {
			s.Log.Infof("Menghapus file backup yang gagal: %s", fullOutputPath)
			os.Remove(fullOutputPath)
		}

		// Tampilkan error detail dari mysqldump
		ui.PrintHeader("ERROR : Mysqldump Gagal dijalankan")

		if writeResult != nil && writeResult.StderrOutput != "" {
			ui.PrintSubHeader("Detail Error dari mysqldump:")
			ui.PrintError(writeResult.StderrOutput)
		}

		errorMsg := fmt.Errorf("gagal menjalankan mysqldump: %w", err)
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
			"combined",
			dbFiltered,
			backupStartTime,
			time.Now(),
			writeResult.SHA256Hash,
			writeResult.MD5Hash,
			fileSize,
		)
		if err := s.WriteBackupMetadata(metadata); err != nil {
			s.Log.Warnf("Gagal menulis metadata file: %v", err)
		}

		// Log checksums
		s.Log.Infof("✓ Checksums: SHA256=%s", writeResult.SHA256Hash[:16]+"...")
		s.Log.Infof("✓ Checksums: MD5=%s", writeResult.MD5Hash[:16]+"...")

		// Verify checksum jika enabled
		if s.Config.Backup.Verification.VerifyAfterWrite {
			s.Log.Info("Memverifikasi checksum backup file...")

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
				s.Log.Errorf("✗ Gagal verifikasi checksum: %v", err)
			} else if verifyResult.Success {
				s.Log.Info("✓ Verifikasi checksum BERHASIL - file backup valid!")
			} else {
				s.Log.Errorf("✗ Verifikasi checksum GAGAL: %s", verifyResult.Error)
			}
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
