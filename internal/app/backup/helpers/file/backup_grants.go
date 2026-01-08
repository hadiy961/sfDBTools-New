package file

import (
	"os"
	"path/filepath"
	"strings"

	"sfdbtools/pkg/compress"
	"sfdbtools/pkg/consts"
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

	if _, err := os.Stat(expectedGrantsPath); err == nil {
		return expectedGrantsPath
	}

	return ""
}
