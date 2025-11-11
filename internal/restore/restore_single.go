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
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"
	"time"

	"github.com/briandowns/spinner"
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

		// Priority 2: Extract dari filename menggunakan fixed pattern
		if targetDB == "" {
			targetDB = extractDatabaseNameFromPattern(sourceFile)

			if targetDB != "" {
				sourceDatabaseName = targetDB
				s.Log.Infof("✓ Target database dari filename: %s", targetDB)
			} else {
				// Pattern tidak match, log warning
				s.Log.Warnf("⚠ Filename tidak sesuai dengan pattern: %s", FixedBackupPattern)
				s.Log.Warnf("  Backup file: %s", filepath.Base(sourceFile))
			}
		}

		// Priority 3: Interactive prompt jika pattern tidak match atau metadata tidak ada
		if targetDB == "" {
			// Check if running in interactive mode (not quiet, has TTY)
			quietMode := helper.GetEnvOrDefault(consts.ENV_QUIET, "false") == "true"
			if !quietMode {
				s.Log.Info("Filename tidak sesuai pattern, gunakan interactive mode untuk input database name...")
				promptedDB, err := s.promptDatabaseName(sourceFile)
				if err != nil {
					return result, fmt.Errorf("gagal mendapatkan database name: %w", err)
				}
				targetDB = promptedDB
				sourceDatabaseName = targetDB
				s.Log.Infof("✓ Target database dari user input: %s", targetDB)
			} else {
				// Quiet mode atau no TTY - tidak bisa interactive
				return result, fmt.Errorf("filename tidak sesuai pattern %s, gunakan flag --target-db (backup file: %s)", FixedBackupPattern, filepath.Base(sourceFile))
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
			sourceDatabaseName = extractDatabaseNameFromPattern(sourceFile)
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

	// Pre-backup before restore (safety backup) - skip jika flag --skip-backup aktif
	if !s.RestoreOptions.SkipBackup {
		s.Log.Info("Creating safety backup before restore...")
		// Validasi database target
		if exists, err := s.isTargetDatabaseExists(ctx, targetDB); err != nil {
			return result, fmt.Errorf("validasi database target gagal: %w", err)
		} else if !exists {
			s.Log.Infof("Database target tidak ada, membuat database baru: %s", targetDB)
			if err := s.Client.CreateDatabase(ctx, targetDB); err != nil {
				return result, fmt.Errorf("gagal create target database: %w", err)
			}
			s.Log.Infof("✓ Database %s berhasil dibuat, skip pre-backup", targetDB)
		} else {
			preBackupFile, err := s.executePreBackup(ctx, targetDB)
			if err != nil {
				return result, fmt.Errorf("gagal create pre-backup: %w", err)
			}
			result.PreBackupFile = preBackupFile
			s.Log.Infof("✓ Safety backup created: %s", preBackupFile)
		}
	}

	// Drop target database jika flag --drop-target aktif
	if s.RestoreOptions.DropTarget {
		// Check if target database exists
		if exists, err := s.isTargetDatabaseExists(ctx, targetDB); err != nil {
			return result, fmt.Errorf("gagal check keberadaan database target: %w", err)
		} else if exists {
			s.Log.Infof("Dropping target database: %s", targetDB)
			if err := s.Client.DropDatabase(ctx, targetDB); err != nil {
				return result, fmt.Errorf("gagal drop target database %s: %w", targetDB, err)
			}
			s.Log.Infof("✓ Database %s berhasil di-drop", targetDB)
		} else {
			s.Log.Infof("Database target %s tidak ada, skip drop", targetDB)
		}
	}

	// Execute restore
	ui.PrintSubHeader("Restore Database ")
	s.Log.Info("Starting database restore...")
	// Create database if not exists
	if exists, err := s.isTargetDatabaseExists(ctx, targetDB); err != nil {
		return result, fmt.Errorf("gagal check keberadaan database target: %w", err)
	} else if !exists {
		s.Log.Infof("Database target tidak ada, membuat database baru: %s", targetDB)
		if err := s.Client.CreateDatabase(ctx, targetDB); err != nil {
			return result, fmt.Errorf("gagal create target database: %w", err)
		}
		s.Log.Infof("✓ Database %s berhasil dibuat", targetDB)
	}

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

		// Error sudah di-log dengan output di executeMysqlRestore(), jangan duplikasi
	} else {
		result.TotalDatabases = 1
		result.SuccessfulRestore = 1
		restoreInfo.Status = "success"

		// Jika ada warning (dari force mode), log ke result juga
		if restoreInfo.Warnings != "" {
			result.Errors = append(result.Errors, "WARNING: "+restoreInfo.Warnings)
			s.Log.Warnf("Restore success dengan warning: %s", restoreInfo.Warnings)
		}
	}

	restoreInfo.Duration = global.FormatDuration(time.Since(startTime))
	result.RestoreInfo = append(result.RestoreInfo, restoreInfo)
	result.TotalTimeTaken = time.Since(startTime)

	return result, nil
}

// restoreSingleDatabase melakukan restore satu database dari backup file
func (s *Service) restoreSingleDatabase(ctx context.Context, sourceFile, targetDB, sourceDatabaseName string) (types.DatabaseRestoreInfo, error) {
	dbName := sourceDatabaseName
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

	// Setup max_statement_time untuk GLOBAL restore (set ke unlimited untuk restore jangka panjang)
	// GLOBAL scope agar mysql CLI juga affected (mysql CLI membuat koneksi terpisah)
	restore, originalMaxStatementTime, err := database.WithGlobalMaxStatementTime(ctx, s.Client, 0)
	if err != nil {
		s.Log.Warnf("Setup GLOBAL max_statement_time gagal: %v", err)
	} else {
		s.Log.Infof("Original GLOBAL max_statement_time: %f detik", originalMaxStatementTime)
		defer func() {
			if rerr := restore(context.Background()); rerr != nil {
				s.Log.Warnf("Gagal mengembalikan GLOBAL max_statement_time: %v", rerr)
			} else {
				s.Log.Info("GLOBAL max_statement_time berhasil dikembalikan.")
			}
		}()
	}

	// Check max_allowed_packet sebelum restore
	maxPacket, err := s.Client.GetMaxAllowedPacket(ctx)
	if err != nil {
		s.Log.Warnf("Gagal mendapatkan max_allowed_packet: %v", err)
	} else {
		s.Log.Infof("max_allowed_packet: %d bytes (%.2f MB)", maxPacket, float64(maxPacket)/1024/1024)
		if maxPacket < 16*1024*1024 { // Less than 16MB
			s.Log.Warnf("⚠ max_allowed_packet kecil (< 16MB), kemungkinan ada packet size issue saat restore")
		}
	}

	// Execute mysql restore menggunakan pipe
	if err := s.executeMysqlRestore(ctx, reader, targetDB, sourceFile, sourceDatabaseName); err != nil {
		s.Log.Debugf("[DEBUG] MySQL restore error detected: %v, Force=%v", err, s.RestoreOptions.Force)
		// Jika force=true, log warning tapi tetap success (dengan warning)
		// Jika force=false, return error (restore gagal)
		if s.RestoreOptions.Force {
			s.Log.Warnf("MySQL restore memiliki error tapi tetap berjalan (--force mode): %v", err)
			info.Warnings = fmt.Sprintf("MySQL restore memiliki error tapi tetap berjalan: %v", err)
		} else {
			s.Log.Errorf("[ERROR] Restore gagal dengan force=false, returning error: %v", err)
			return info, fmt.Errorf("gagal restore database: %w", err)
		}
	} else {
		s.Log.Debugf("[DEBUG] MySQL restore completed without error")
	}

	info.Verified = s.RestoreOptions.VerifyChecksum
	s.Log.Infof("✓ Database %s berhasil di-restore", targetDB)

	return info, nil
}

// executeMysqlRestore menjalankan mysql command untuk restore dari reader
func (s *Service) executeMysqlRestore(ctx context.Context, reader io.Reader, targetDB, sourceFile, sourceDatabaseName string) error {
	// Build mysql command args - hanya tambahkan force jika true
	args := []string{
		fmt.Sprintf("--host=%s", s.TargetProfile.DBInfo.Host),
		fmt.Sprintf("--port=%d", s.TargetProfile.DBInfo.Port),
		fmt.Sprintf("--user=%s", s.TargetProfile.DBInfo.User),
		fmt.Sprintf("--password=%s", s.TargetProfile.DBInfo.Password),
	}

	// Tambahkan force flag hanya jika true (untuk continue on errors)
	if s.RestoreOptions.Force {
		args = append(args, "--force")
	}

	// Tambahkan target database
	args = append(args, targetDB)

	// Start spinner dengan elapsed time
	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	spin.Suffix = " Melakukan restore database..."
	spin.Start()
	defer spin.Stop()

	// Goroutine untuk update spinner dengan elapsed time
	startTime := time.Now()
	done := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				elapsed := time.Since(startTime)
				spin.Suffix = fmt.Sprintf(" Melakukan restore database... (%s)", global.FormatDuration(elapsed))
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	cmd := exec.CommandContext(ctx, "mysql", args...)
	cmd.Stdin = reader

	// Capture output
	output, err := cmd.CombinedOutput()
	done <- true // Stop elapsed time updater

	if err != nil {
		spin.Stop()
		fmt.Println()
		s.Log.Error("MySQL restore gagal, lihat log untuk detail")
		// Log error detail ke file terpisah (jangan di output) dan tampilkan di mana logsnya
		logFile := s.ErrorLog.LogWithOutput(map[string]interface{}{
			"database": sourceDatabaseName,
			"source":   sourceFile,
			"target":   targetDB,
			"force":    s.RestoreOptions.Force,
		}, string(output), err)
		if logFile != "" {
			s.Log.Infof("Error details tersimpan di: %s", logFile)
		}
		// Return error, logic force handling di level caller
		return fmt.Errorf("mysql restore failed: %w", err)
	}

	if len(output) > 0 {
		s.Log.Debugf("MySQL output: %s", string(output))
	}

	return nil
}
