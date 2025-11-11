package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/pkg/helper"
	"sort"
	"strings"
)

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
