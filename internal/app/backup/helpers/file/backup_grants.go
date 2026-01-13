package file

import (
	"path/filepath"
	"strings"

	"sfdbtools/internal/shared/compress"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/fsops"
)

// GenerateGrantsFilename generate expected grants filename dari backup filename.
func GenerateGrantsFilename(backupFilename string) string {
	nameWithoutExt := backupFilename
	nameWithoutExt = strings.TrimSuffix(nameWithoutExt, consts.ExtEnc)
	for _, ext := range compress.SupportedCompressionExtensions() {
		nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ext)
	}
	nameWithoutExt = strings.TrimSuffix(nameWithoutExt, consts.ExtSQL)

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
