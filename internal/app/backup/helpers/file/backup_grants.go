package file

import (
	"path/filepath"

	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/fsops"
)

// GenerateGrantsFilename generate expected grants filename dari backup filename.
func GenerateGrantsFilename(backupFilename string) string {
	nameWithoutExt, _ := ExtractFileExtensions(backupFilename)
	return nameWithoutExt + consts.UsersSQLSuffix
}

// AutoDetectGrantsFile auto-detect file grants berdasarkan backup file.
func AutoDetectGrantsFile(backupFile string) string {
	dir := filepath.Dir(backupFile)
	basename := filepath.Base(backupFile)

	expectedGrantsFile := GenerateGrantsFilename(basename)
	expectedGrantsPath := filepath.Join(dir, expectedGrantsFile)

	if err := fsops.ValidateFileExists(expectedGrantsPath); err == nil {
		return expectedGrantsPath
	}

	return ""
}
