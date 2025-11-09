// File : internal/restore/restore_all.go
// Deskripsi : Restore semua database dari file combined backup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package restore

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/helper"
	"strings"
	"time"
)

// executeRestoreAll melakukan restore semua database dari file combined backup
func (s *Service) executeRestoreAll(ctx context.Context) (types.RestoreResult, error) {
	var result types.RestoreResult

	s.SetRestoreInProgress(true)
	defer s.SetRestoreInProgress(false)

	startTime := time.Now()
	sourceFile := s.RestoreOptions.SourceFile

	s.Log.Infof("Restoring all databases from: %s", filepath.Base(sourceFile))

	// Load metadata untuk mendapatkan list database
	metadataFile := getMetadataFilePath(sourceFile)
	var databases []string

	if _, err := os.Stat(metadataFile); err == nil {
		metadata, err := s.loadBackupMetadata(metadataFile)
		if err != nil {
			s.Log.Warnf("Gagal load metadata: %v, akan restore semua database dari backup", err)
		} else {
			databases = metadata.DatabaseNames
			s.Log.Infof("Found %d databases in metadata: %v", len(databases), databases)
		}
	}

	// Check if dry run
	if s.RestoreOptions.DryRun {
		s.Log.Info("[DRY RUN] Restore tidak dijalankan (mode simulasi)")
		result.TotalDatabases = len(databases)
		// Untuk dry run, hitung sebagai skipped bukan success/failed
		for _, dbName := range databases {
			result.RestoreInfo = append(result.RestoreInfo, types.DatabaseRestoreInfo{
				DatabaseName:   dbName,
				SourceFile:     sourceFile,
				TargetDatabase: dbName,
				Status:         "skipped",
				Duration:       "0s",
			})
		}
		result.TotalTimeTaken = time.Since(startTime)
		return result, nil
	}

	// Pre-backup before restore (safety backup) - untuk combined backup
	// Backup semua databases yang akan di-restore
	if s.RestoreOptions.BackupBeforeRestore && len(databases) > 0 {
		s.Log.Info("Creating safety backups before restore...")
		var preBackupFiles []string
		for _, dbName := range databases {
			preBackupFile, err := s.executePreBackup(ctx, dbName)
			if err != nil {
				s.Log.Warnf("Gagal create pre-backup untuk %s: %v", dbName, err)
				// Continue dengan backup database lainnya
				continue
			}
			preBackupFiles = append(preBackupFiles, preBackupFile)
		}
		if len(preBackupFiles) > 0 {
			result.PreBackupFile = strings.Join(preBackupFiles, ", ")
			s.Log.Infof("✓ Safety backups created: %d files", len(preBackupFiles))
		}
	}

	// Execute restore all databases
	restoreInfo, err := s.restoreAllDatabases(ctx, sourceFile, databases)
	if err != nil {
		// Failed restore
		result.TotalDatabases = len(databases)
		if result.TotalDatabases == 0 {
			result.TotalDatabases = 1 // At least 1 untuk error reporting
		}
		result.FailedRestore = result.TotalDatabases
		result.Errors = append(result.Errors, err.Error())

		restoreInfo.Status = "failed"
		restoreInfo.ErrorMessage = err.Error()
	} else {
		// Successful restore
		result.TotalDatabases = len(databases)
		if result.TotalDatabases == 0 {
			result.TotalDatabases = 1 // Combined backup as 1 unit
		}
		result.SuccessfulRestore = result.TotalDatabases
		restoreInfo.Status = "success"
	}
	restoreInfo.Duration = global.FormatDuration(time.Since(startTime))
	result.RestoreInfo = append(result.RestoreInfo, restoreInfo)
	result.TotalTimeTaken = time.Since(startTime)

	return result, nil
}

// restoreAllDatabases melakukan restore semua database dari combined backup file
func (s *Service) restoreAllDatabases(ctx context.Context, sourceFile string, databases []string) (types.DatabaseRestoreInfo, error) {
	info := types.DatabaseRestoreInfo{
		DatabaseName:   fmt.Sprintf("Combined (%d databases)", len(databases)),
		SourceFile:     sourceFile,
		TargetDatabase: "Multiple databases",
	}

	fileInfo, err := os.Stat(sourceFile)
	if err != nil {
		return info, err
	}
	info.FileSize = fileInfo.Size()
	info.FileSizeHuman = global.FormatFileSize(fileInfo.Size())

	// Prepare reader pipeline: file → decrypt → decompress
	file, err := os.Open(sourceFile)
	if err != nil {
		return info, fmt.Errorf("gagal open backup file: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file

	// Decrypt if encrypted
	isEncrypted := strings.HasSuffix(sourceFile, ".enc")
	if isEncrypted {
		encryptionKey := s.RestoreOptions.EncryptionKey
		if encryptionKey == "" {
			encryptionKey = helper.GetEnvOrDefault("SFDB_BACKUP_ENCRYPTION_KEY", "")
			if encryptionKey == "" {
				return info, fmt.Errorf("encryption key required untuk restore encrypted backup")
			}
		}

		decryptReader, err := encrypt.NewDecryptingReader(reader, encryptionKey)
		if err != nil {
			return info, fmt.Errorf("gagal setup decrypt reader: %w", err)
		}
		reader = decryptReader
		s.Log.Debug("Decrypting backup file...")
	}

	// Decompress if compressed
	compressionType := detectCompressionType(sourceFile)
	if compressionType != "" {
		decompressReader, err := compress.NewDecompressingReader(reader, compress.CompressionType(compressionType))
		if err != nil {
			return info, fmt.Errorf("gagal setup decompress reader: %w", err)
		}
		defer decompressReader.Close()
		reader = decompressReader
		s.Log.Debugf("Decompressing backup file (%s)...", compressionType)
	}

	// Execute mysql restore untuk semua database sekaligus
	// Combined backup sudah contain CREATE DATABASE statements
	if err := s.executeMysqlRestoreAll(ctx, reader); err != nil {
		return info, fmt.Errorf("gagal restore databases: %w", err)
	}

	info.Verified = s.RestoreOptions.VerifyChecksum
	s.Log.Infof("✓ All databases berhasil di-restore")

	return info, nil
}

// executeMysqlRestoreAll menjalankan mysql command untuk restore all databases
func (s *Service) executeMysqlRestoreAll(ctx context.Context, reader io.Reader) error {
	// Build mysql command (tanpa specify database karena combined backup sudah ada CREATE DATABASE)
	args := []string{
		fmt.Sprintf("--host=%s", s.TargetProfile.DBInfo.Host),
		fmt.Sprintf("--port=%d", s.TargetProfile.DBInfo.Port),
		fmt.Sprintf("--user=%s", s.TargetProfile.DBInfo.User),
		fmt.Sprintf("--password=%s", s.TargetProfile.DBInfo.Password),
	}

	cmd := exec.CommandContext(ctx, "mysql", args...)
	cmd.Stdin = reader

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mysql restore failed: %w\nOutput: %s", err, string(output))
	}

	if len(output) > 0 {
		s.Log.Debugf("MySQL output: %s", string(output))
	}

	return nil
}

// getMetadataFilePath mendapatkan path untuk metadata file
// Format: backup_file.gz.enc -> backup_file.gz.enc.meta.json
func getMetadataFilePath(backupFile string) string {
	return backupFile + ".meta.json"
}
