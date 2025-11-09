// File : internal/restore/restore_single.go
// Deskripsi : Restore database dari satu file backup terpisah
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
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/helper"
	"strings"
	"time"
)

// executeRestoreSingle melakukan restore database dari satu file backup terpisah
func (s *Service) executeRestoreSingle(ctx context.Context) (types.RestoreResult, error) {
	var result types.RestoreResult

	s.SetRestoreInProgress(true)
	defer s.SetRestoreInProgress(false)

	startTime := time.Now()
	sourceFile := s.RestoreOptions.SourceFile

	// Extract database name jika target DB tidak specified
	// Priority 1: Load dari metadata file
	// Priority 2: Extract dari filename (fallback)
	targetDB := s.RestoreOptions.TargetDB
	sourceDatabaseName := ""

	if targetDB == "" {
		// Priority 1: Load database name dari metadata
		metadataFile := sourceFile + ".meta.json"
		if _, err := os.Stat(metadataFile); err == nil {
			// Metadata exists, load database name from it
			metadata, err := s.loadBackupMetadata(metadataFile)
			if err != nil {
				s.Log.Warnf("Gagal load metadata file: %v", err)
			} else if len(metadata.DatabaseNames) > 0 {
				sourceDatabaseName = metadata.DatabaseNames[0]
				targetDB = sourceDatabaseName
				s.Log.Infof("✓ Target database dari metadata: %s", targetDB)
			}
		}

		// Priority 2: Extract dari filename menggunakan pattern dari config
		if targetDB == "" {
			namePattern := s.Config.Backup.Output.NamePattern
			targetDB = extractDatabaseNameFromPattern(sourceFile, namePattern)

			if targetDB != "" {
				sourceDatabaseName = targetDB
				// Log berbeda tergantung apakah pakai pattern atau legacy
				if namePattern != "" {
					s.Log.Infof("✓ Target database dari filename (pattern: %s): %s", namePattern, targetDB)
				} else {
					s.Log.Infof("✓ Target database dari filename: %s (FALLBACK - pattern tidak tersedia)", targetDB)
				}
			}
		}

		// Priority 3: Interactive prompt sebagai last resort
		if targetDB == "" {
			// Check if running in interactive mode (not quiet, has TTY)
			quietMode := helper.GetEnvOrDefault(consts.ENV_QUIET, "false") == "true"
			if !quietMode {
				s.Log.Info("Menggunakan interactive mode untuk input database name...")
				promptedDB, err := s.promptDatabaseName(sourceFile)
				if err != nil {
					return result, fmt.Errorf("gagal mendapatkan database name: %w", err)
				}
				targetDB = promptedDB
				sourceDatabaseName = targetDB
				s.Log.Infof("✓ Target database dari user input: %s", targetDB)
			} else {
				// Quiet mode atau no TTY - tidak bisa interactive
				return result, fmt.Errorf("gagal extract database name dari filename, gunakan flag --target-db (backup file: %s)", sourceFile)
			}
		}
	} else {
		// User specified target DB, but we still need source DB name for display
		metadataFile := sourceFile + ".meta.json"
		if _, err := os.Stat(metadataFile); err == nil {
			metadata, err := s.loadBackupMetadata(metadataFile)
			if err == nil && len(metadata.DatabaseNames) > 0 {
				sourceDatabaseName = metadata.DatabaseNames[0]
			}
		}
		if sourceDatabaseName == "" {
			namePattern := s.Config.Backup.Output.NamePattern
			sourceDatabaseName = extractDatabaseNameFromPattern(sourceFile, namePattern)
		}
	}

	if sourceDatabaseName != "" && targetDB != sourceDatabaseName {
		s.Log.Infof("Restoring database: %s -> %s from %s", sourceDatabaseName, targetDB, filepath.Base(sourceFile))
	} else {
		s.Log.Infof("Restoring database: %s from %s", targetDB, filepath.Base(sourceFile))
	}

	// Check if dry run
	if s.RestoreOptions.DryRun {
		s.Log.Info("[DRY RUN] Restore tidak dijalankan (mode simulasi)")
		result.TotalDatabases = 1
		dbName := sourceDatabaseName
		if dbName == "" {
			dbName = targetDB
		}
		result.RestoreInfo = append(result.RestoreInfo, types.DatabaseRestoreInfo{
			DatabaseName:   dbName,
			SourceFile:     sourceFile,
			TargetDatabase: targetDB,
			Status:         "skipped",
			Duration:       "0s",
		})
		return result, nil
	}

	// Pre-backup before restore (safety backup)
	if s.RestoreOptions.BackupBeforeRestore {
		s.Log.Info("Creating safety backup before restore...")
		preBackupFile, err := s.executePreBackup(ctx, targetDB)
		if err != nil {
			return result, fmt.Errorf("gagal create pre-backup: %w", err)
		}
		result.PreBackupFile = preBackupFile
		s.Log.Infof("✓ Safety backup created: %s", preBackupFile)
	}

	// Execute restore
	restoreInfo, err := s.restoreSingleDatabase(ctx, sourceFile, targetDB, sourceDatabaseName)
	if err != nil {
		result.TotalDatabases = 1
		result.FailedRestore = 1
		result.FailedDatabases = map[string]string{
			targetDB: err.Error(),
		}
		result.Errors = append(result.Errors, err.Error())

		restoreInfo.Status = "failed"
		restoreInfo.ErrorMessage = err.Error()
	} else {
		result.TotalDatabases = 1
		result.SuccessfulRestore = 1
		restoreInfo.Status = "success"
	}

	restoreInfo.Duration = global.FormatDuration(time.Since(startTime))
	result.RestoreInfo = append(result.RestoreInfo, restoreInfo)
	result.TotalTimeTaken = time.Since(startTime)

	return result, nil
}

// restoreSingleDatabase melakukan restore satu database dari backup file
func (s *Service) restoreSingleDatabase(ctx context.Context, sourceFile, targetDB, sourceDatabaseName string) (types.DatabaseRestoreInfo, error) {
	dbName := sourceDatabaseName
	if dbName == "" {
		namePattern := s.Config.Backup.Output.NamePattern
		dbName = extractDatabaseNameFromPattern(sourceFile, namePattern)
	}

	info := types.DatabaseRestoreInfo{
		DatabaseName:   dbName,
		SourceFile:     sourceFile,
		TargetDatabase: targetDB,
	}

	fileInfo, err := os.Stat(sourceFile)
	if err != nil {
		return info, err
	}
	info.FileSize = fileInfo.Size()
	info.FileSizeHuman = global.FormatFileSize(fileInfo.Size())

	// Check if target database exists, create if not
	exists, err := s.Client.DatabaseExists(ctx, targetDB)
	if err != nil {
		return info, fmt.Errorf("gagal check database existence: %w", err)
	}

	if !exists {
		s.Log.Infof("Creating target database: %s", targetDB)
		if err := s.Client.CreateDatabase(ctx, targetDB); err != nil {
			return info, fmt.Errorf("gagal create target database: %w", err)
		}
	}

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

	// Execute mysql restore menggunakan pipe
	if err := s.executeMysqlRestore(ctx, reader, targetDB); err != nil {
		return info, fmt.Errorf("gagal restore database: %w", err)
	}

	info.Verified = s.RestoreOptions.VerifyChecksum
	s.Log.Infof("✓ Database %s berhasil di-restore", targetDB)

	return info, nil
}

// executeMysqlRestore menjalankan mysql command untuk restore dari reader
func (s *Service) executeMysqlRestore(ctx context.Context, reader io.Reader, targetDB string) error {
	// Build mysql command
	args := []string{
		fmt.Sprintf("--host=%s", s.TargetProfile.DBInfo.Host),
		fmt.Sprintf("--port=%d", s.TargetProfile.DBInfo.Port),
		fmt.Sprintf("--user=%s", s.TargetProfile.DBInfo.User),
		fmt.Sprintf("--password=%s", s.TargetProfile.DBInfo.Password),
		targetDB,
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
