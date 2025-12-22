package helper

import (
	"os"
	"path/filepath"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
	"strings"
)

// ExtractDatabaseNameFromFile mengekstrak nama database dari nama file backup
func ExtractDatabaseNameFromFile(filePath string) string {
	if filePath == "" {
		return ""
	}

	fileName := filepath.Base(filePath)

	// Remove extensions
	fileName = strings.TrimSuffix(fileName, consts.ExtEnc)
	fileName = strings.TrimSuffix(fileName, consts.ExtGzip)
	fileName = strings.TrimSuffix(fileName, consts.ExtXz)
	fileName = strings.TrimSuffix(fileName, consts.ExtZstd)
	fileName = strings.TrimSuffix(fileName, consts.ExtZlib)
	fileName = strings.TrimSuffix(fileName, consts.ExtSQL)

	// Split by underscore
	parts := strings.Split(fileName, "_")
	if len(parts) == 0 {
		return fileName
	}

	// Find where timestamp/hostname pattern starts
	cutIndex := len(parts)

	for i := 0; i < len(parts); i++ {
		part := parts[i]

		// Check if this is timestamp date (8 digits: YYYYMMDD)
		if len(part) == 8 && isAllDigits(part) {
			cutIndex = i
			break
		}

		// Check if this is timestamp time (6 digits: HHMMSS)
		if len(part) == 6 && isAllDigits(part) && i > 0 {
			if len(parts[i-1]) == 8 && isAllDigits(parts[i-1]) {
				cutIndex = i - 1
				break
			}
		}

		// Check for common suffixes
		if part == "backup" || part == "dump" || part == "pre" || part == "restore" {
			cutIndex = i
			break
		}

		// Check if this is hostname pattern
		if strings.Contains(part, "-") && !isAllDigits(part) {
			cutIndex = i
			break
		}
	}

	if cutIndex > 0 {
		return strings.Join(parts[:cutIndex], "_")
	}

	return fileName
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// IsValidDatabaseName validates MySQL/MariaDB database name
func IsValidDatabaseName(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

// ListBackupFilesInDirectory membaca daftar file backup di direktori
func ListBackupFilesInDirectory(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.Contains(name, consts.ExtSQL) {
			files = append(files, name)
		}
	}

	return files, nil
}

// GenerateGrantsFilename generate expected grants filename dari backup filename
func GenerateGrantsFilename(backupFilename string) string {
	nameWithoutExt := backupFilename
	nameWithoutExt = strings.TrimSuffix(nameWithoutExt, consts.ExtEnc)
	for _, ext := range compress.SupportedCompressionExtensions() {
		nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ext)
	}
	nameWithoutExt = strings.TrimSuffix(nameWithoutExt, consts.ExtSQL)

	return nameWithoutExt + consts.UsersSQLSuffix
}

// AutoDetectGrantsFile auto-detect file grants berdasarkan backup file
func AutoDetectGrantsFile(backupFile string) string {
	dir := filepath.Dir(backupFile)
	basename := filepath.Base(backupFile)

	expectedGrantsFile := GenerateGrantsFilename(basename)
	expectedGrantsPath := filepath.Join(dir, expectedGrantsFile)

	if _, err := os.Stat(expectedGrantsPath); err == nil {
		return expectedGrantsPath
	}

	return ""
}
