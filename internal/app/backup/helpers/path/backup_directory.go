package path

import (
	"fmt"
	"os"
	"path/filepath"

	"sfdbtools/internal/shared/compress"
	"sfdbtools/internal/shared/consts"
)

const (
	// MaxPathLength adalah limit maksimum path di Linux (PATH_MAX)
	MaxPathLength = 4096
)

// ValidateBackupDirectory melakukan pre-flight check untuk backup directory.
// Fungsi ini memvalidasi bahwa directory exists, adalah directory (bukan file),
// dan writable. Jika createIfMissing=true, akan mencoba membuat directory jika belum ada.
func ValidateBackupDirectory(baseDir string, createIfMissing bool) error {
	if baseDir == "" {
		return fmt.Errorf("backup directory path is empty")
	}

	// Check if exists
	info, err := os.Stat(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			if createIfMissing {
				// Try to create with proper permissions (0755 = rwxr-xr-x)
				if err := os.MkdirAll(baseDir, 0755); err != nil {
					return fmt.Errorf("cannot create backup directory %s: %w", baseDir, err)
				}
				// Verify creation successful
				if _, err := os.Stat(baseDir); err != nil {
					return fmt.Errorf("backup directory created but cannot access: %s", baseDir)
				}
				return nil
			}
			return fmt.Errorf("backup directory does not exist: %s", baseDir)
		}
		return fmt.Errorf("cannot access backup directory %s: %w", baseDir, err)
	}

	// Check if directory (not file)
	if !info.IsDir() {
		return fmt.Errorf("backup path is not a directory: %s", baseDir)
	}

	// Check writable (create temp file)
	testFile := filepath.Join(baseDir, ".sfdbtools-write-test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		return fmt.Errorf("backup directory not writable: %s", baseDir)
	}
	os.Remove(testFile) // cleanup test file

	return nil
}

// ValidatePathLength memvalidasi bahwa path tidak melebihi filesystem limit.
// Linux PATH_MAX adalah 4096 bytes.
func ValidatePathLength(path string) error {
	if len(path) > MaxPathLength {
		// Truncate path untuk error message agar tidak spam log
		truncated := path
		if len(path) > 100 {
			truncated = path[:100] + "..."
		}
		return fmt.Errorf("backup path too long (%d chars, max %d): %s",
			len(path), MaxPathLength, truncated)
	}
	return nil
}

// EnsureBackupDirectory memastikan directory exists dengan atomic creation.
// Thread-safe untuk concurrent backup operations.
func EnsureBackupDirectory(path string) error {
	if path == "" {
		return fmt.Errorf("backup directory path is empty")
	}

	// MkdirAll is atomic and thread-safe
	// os.IsExist check diperlukan karena MkdirAll bisa return error jika dir sudah ada
	// tapi dengan permission berbeda
	if err := os.MkdirAll(path, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	return nil
}

func GenerateBackupDirectory(baseDir string, structurePattern string, hostname string) (string, error) {
	// Validate base directory early (fail fast)
	if err := ValidateBackupDirectory(baseDir, true); err != nil {
		return "", fmt.Errorf("backup directory validation failed: %w", err)
	}
	replacer, err := NewPathPatternReplacer("", hostname, compress.CompressionType(consts.CompressionTypeNone), false, false)
	if err != nil {
		return "", fmt.Errorf("gagal membuat pattern replacer: %w", err)
	}

	subPath := replacer.ReplacePattern(structurePattern, true)
	fullPath := filepath.Join(baseDir, subPath)

	// Validate path length (before directory creation)
	if err := ValidatePathLength(fullPath); err != nil {
		return "", err
	}

	// Ensure directory exists (atomic creation)
	if err := EnsureBackupDirectory(fullPath); err != nil {
		return "", err
	}

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
