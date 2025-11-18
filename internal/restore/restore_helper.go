package restore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/profilehelper"
	"sort"
	"strings"
)

// DatabaseNameResolution berisi hasil resolusi database name
type DatabaseNameResolution struct {
	TargetDB     string // Database name yang akan digunakan sebagai target
	SourceDB     string // Database name dari source (untuk display/logging)
	ResolvedFrom string // "filename", "user_input", "flag"
}

// resolveDatabaseName mendapatkan database name dengan priority:
// 1. User-specified target DB (dari flag --target-db)
// 2. Extract dari filename pattern
// 3. Interactive prompt (jika tidak quiet mode)
func (s *Service) resolveDatabaseName(sourceFile string, userSpecifiedTargetDB string) (*DatabaseNameResolution, error) {
	result := &DatabaseNameResolution{
		TargetDB: userSpecifiedTargetDB,
	}

	// Priority 1: User specified target DB via flag
	if userSpecifiedTargetDB != "" {
		result.ResolvedFrom = "flag"
		// Masih perlu resolve source DB untuk display
		result.SourceDB = s.tryGetSourceDatabaseName(sourceFile)
		if result.SourceDB == "" {
			result.SourceDB = userSpecifiedTargetDB // Fallback
		}
		s.Log.Debugf("Target DB from flag: %s", result.TargetDB)
		return result, nil
	}

	// Priority 2: Extract dari filename pattern
	dbName := extractDatabaseNameFromPattern(sourceFile)
	if dbName != "" {
		result.TargetDB = dbName
		result.SourceDB = dbName
		result.ResolvedFrom = "filename"
		s.Log.Infof("✓ Target database dari filename: %s", result.TargetDB)
		return result, nil
	}

	// Pattern tidak match
	s.Log.Warnf("⚠ Filename tidak sesuai dengan pattern: %s", FixedBackupPattern)
	s.Log.Warnf("  Backup file: %s", filepath.Base(sourceFile))

	// Priority 3: Interactive prompt jika tidak quiet mode
	quietMode := helper.GetEnvOrDefault(consts.ENV_QUIET, "false") == "true"
	if !quietMode {
		s.Log.Info("Filename tidak sesuai pattern, gunakan interactive mode untuk input database name...")
		promptedDB, err := s.promptDatabaseName(sourceFile)
		if err != nil {
			return nil, fmt.Errorf("gagal mendapatkan database name: %w", err)
		}
		result.TargetDB = promptedDB
		result.SourceDB = promptedDB
		result.ResolvedFrom = "user_input"
		s.Log.Infof("✓ Target database dari user input: %s", result.TargetDB)
		return result, nil
	}

	// Quiet mode atau no TTY - tidak bisa interactive
	return nil, fmt.Errorf("filename tidak sesuai pattern %s, gunakan flag --target-db (backup file: %s)",
		FixedBackupPattern, filepath.Base(sourceFile))
}

// tryGetSourceDatabaseName mencoba mendapatkan source database name untuk display/logging
// Tidak error jika gagal, return empty string
func (s *Service) tryGetSourceDatabaseName(sourceFile string) string {
	// Try filename pattern
	return extractDatabaseNameFromPattern(sourceFile)
}

// getFileInfo mendapatkan informasi file untuk DatabaseRestoreInfo
func getFileInfo(sourceFile string) (fileSize int64, fileSizeHuman string) {
	fileInfo, err := os.Stat(sourceFile)
	if err == nil {
		fileSize = fileInfo.Size()
		fileSizeHuman = global.FormatFileSize(fileSize)
	}
	return
}

// selectLatestBackupFiles memilih file backup terbaru untuk setiap database
// Jika ada multiple files untuk satu database, pilih yang terbaru berdasarkan ModTime
func (s *Service) SelectLatestBackupFiles(backupFiles []BackupFileInfo) []BackupFileInfo {
	// Group files by database name
	dbFilesMap := make(map[string][]BackupFileInfo)

	for _, fileInfo := range backupFiles {
		dbFilesMap[fileInfo.DatabaseName] = append(dbFilesMap[fileInfo.DatabaseName], fileInfo)
	}

	// Select latest file for each database
	var latestFiles []BackupFileInfo

	for dbName, files := range dbFilesMap {
		if len(files) == 1 {
			latestFiles = append(latestFiles, files[0])
			continue
		}

		// Multiple files for same database, sort by ModTime and pick latest
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime.After(files[j].ModTime)
		})

		latestFile := files[0]
		latestFiles = append(latestFiles, latestFile)

		// Log info about duplicate files
		s.Log.Infof("Database %s memiliki %d backup files, menggunakan yang terbaru:", dbName, len(files))
		s.Log.Infof("  ✓ %s (%s)", filepath.Base(latestFile.FilePath), latestFile.ModTime.Format("2006-01-02 15:04:05"))

		if len(files) > 1 {
			for i := 1; i < len(files); i++ {
				s.Log.Infof("  • %s (%s) - skipped",
					filepath.Base(files[i].FilePath),
					files[i].ModTime.Format("2006-01-02 15:04:05"))
			}
		}
	}

	// Sort by database name for consistent output
	sort.Slice(latestFiles, func(i, j int) bool {
		return strings.ToLower(latestFiles[i].DatabaseName) < strings.ToLower(latestFiles[j].DatabaseName)
	})

	return latestFiles
}

// scanBackupFiles membaca direktori dan mengidentifikasi semua backup files
func (s *Service) scanBackupFiles(sourceDir string) ([]BackupFileInfo, error) {
	var backupFiles []BackupFileInfo

	// Validate direktori exists
	dirInfo, err := os.Stat(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("direktori tidak ditemukan: %w", err)
	}
	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("path bukan direktori: %s", sourceDir)
	}

	// Walk direktori untuk mencari backup files
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.Log.Warnf("Error walking path %s: %v", path, err)
			return nil // Continue walking
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is backup file
		if !helper.IsBackupFile(path) {
			s.Log.Debugf("Skipping non-backup file: %s", filepath.Base(path))
			return nil
		}

		// Extract database name dari filename
		dbName := extractDatabaseNameFromPattern(path)
		if dbName == "" {
			s.Log.Warnf("⚠ File tidak sesuai pattern, skip: %s", filepath.Base(path))
			return nil
		}

		backupFiles = append(backupFiles, BackupFileInfo{
			FilePath:     path,
			DatabaseName: dbName,
			ModTime:      info.ModTime(),
			FileSize:     info.Size(),
		})

		s.Log.Debugf("Found backup file: %s -> database: %s", filepath.Base(path), dbName)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	return backupFiles, nil
}

// ensureValidConnection memastikan koneksi database valid, reconnect jika perlu
func (s *Service) ensureValidConnection(ctx context.Context) error {
	// Cek apakah client ada
	if s.Client == nil {
		s.Log.Warn("Database client tidak tersedia, membuat koneksi baru...")
		client, err := profilehelper.ConnectWithProfile(s.TargetProfile, "mysql")
		if err != nil {
			return fmt.Errorf("gagal membuat koneksi database: %w", err)
		}
		s.Client = client
		return nil
	}

	// Ping untuk validasi koneksi
	if err := s.Client.Ping(ctx); err != nil {
		s.Log.Warnf("Koneksi database tidak valid (%v), reconnecting...", err)

		// Close koneksi lama (ignore error)
		s.Client.Close()

		// Buat koneksi baru
		client, err := profilehelper.ConnectWithProfile(s.TargetProfile, "mysql")
		if err != nil {
			return fmt.Errorf("gagal reconnect database: %w", err)
		}
		s.Client = client
		s.Log.Info("✓ Koneksi database berhasil di-restore")
	}

	return nil
}
