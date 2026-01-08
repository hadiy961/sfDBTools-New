package path

import (
	"fmt"
	"path/filepath"

	"sfdbtools/internal/shared/compress"
	"sfdbtools/internal/shared/consts"
)

func GenerateBackupDirectory(baseDir string, structurePattern string, hostname string) (string, error) {
	replacer, err := NewPathPatternReplacer("", hostname, compress.CompressionType(consts.CompressionTypeNone), false, false)
	if err != nil {
		return "", fmt.Errorf("gagal membuat pattern replacer: %w", err)
	}

	subPath := replacer.ReplacePattern(structurePattern, true)
	fullPath := filepath.Join(baseDir, subPath)

	return fullPath, nil
}

// Note: filenamePattern parameter diabaikan karena menggunakan fixed pattern.
func GenerateFullBackupPath(baseDir string, structurePattern string, filenamePattern string, database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool, excludeData bool) (string, error) {
	dir, err := GenerateBackupDirectory(baseDir, structurePattern, hostname)
	if err != nil {
		return "", err
	}

	filename, err := GenerateBackupFilename(database, mode, hostname, compressionType, encrypted, excludeData)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(dir, filename)
	return fullPath, nil
}
