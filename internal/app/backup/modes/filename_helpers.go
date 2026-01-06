// File : internal/backup/modes/filename_helpers.go
// Deskripsi : Helper penamaan file untuk mode backup (custom base name + auto ekstensi)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package modes

import (
	"path/filepath"
	"strings"

	"sfdbtools/pkg/consts"
)

func applyCustomBaseFilename(defaultFilename string, customBase string) string {
	customBase = strings.TrimSpace(customBase)
	if customBase == "" {
		return defaultFilename
	}
	// Jika user sudah memasukkan .sql (atau full filename), biarkan apa adanya.
	if strings.Contains(customBase, ".sql") {
		return customBase
	}

	ext := ""
	if defaultFilename != "" && defaultFilename != consts.FilenameGenerateErrorPlaceholder {
		if idx := strings.Index(defaultFilename, ".sql"); idx >= 0 {
			ext = defaultFilename[idx:]
		} else {
			ext = filepath.Ext(defaultFilename)
		}
	}
	return customBase + ext
}

func getCompanionSuffix(primaryDBName string, dbName string) (string, bool) {
	if primaryDBName == "" || dbName == "" {
		return "", false
	}
	if !strings.HasPrefix(dbName, primaryDBName) {
		return "", false
	}
	remainder := strings.TrimPrefix(dbName, primaryDBName)
	if remainder == "" {
		return "", false
	}
	allowed := map[string]bool{
		"_dmart":   true,
		"_temp":    true,
		"_archive": true,
	}
	if !allowed[remainder] {
		return "", false
	}
	return remainder, true
}

func insertSuffixBeforeSQLExt(filename string, suffix string) string {
	filename = strings.TrimSpace(filename)
	suffix = strings.TrimSpace(suffix)
	if filename == "" || suffix == "" {
		return filename
	}
	if idx := strings.Index(filename, ".sql"); idx >= 0 {
		return filename[:idx] + suffix + filename[idx:]
	}
	return filename + suffix
}
