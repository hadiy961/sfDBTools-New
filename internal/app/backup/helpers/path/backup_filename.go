package path

import (
	"fmt"
	"strings"
	"time"

	"sfdbtools/internal/shared/compress"
	"sfdbtools/internal/shared/consts"
)

func GenerateBackupFilename(database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool, excludeData bool) (string, error) {
	return GenerateBackupFilenameWithCount(database, mode, hostname, compressionType, encrypted, 0, excludeData)
}

func GenerateBackupFilenameWithCount(database string, mode string, hostname string, compressionType compress.CompressionType, encrypted bool, dbCount int, excludeData bool) (string, error) {
	if mode == "separated" && database == "" {
		return "", fmt.Errorf("database name tidak boleh kosong untuk mode separated")
	}

	if mode == "combined" || mode == "all" {
		if dbCount > 0 {
			prefix := mode
			if excludeData {
				prefix += "_nodata"
			}
			timestamp := time.Now().Format("20060102_150405")
			database = fmt.Sprintf("%s_%s_%s_%ddb", prefix, hostname, timestamp, dbCount)
		} else if database == "" {
			database = "all_databases"
		}
	} else if excludeData {
		database += "_nodata"
	}

	replacer, err := NewPathPatternReplacer(database, hostname, compressionType, encrypted, true)
	if err != nil {
		return "", fmt.Errorf("gagal membuat pattern replacer: %w", err)
	}

	var filename string
	if (mode == "combined" || mode == "all") && dbCount > 0 {
		filename = replacer.ReplacePattern("{database}")
	} else {
		filename = replacer.ReplacePattern(consts.FixedBackupPattern)
	}

	if filename == "" || filename == consts.ExtSQL || strings.HasPrefix(filename, ".") {
		return "", fmt.Errorf("hasil generate filename tidak valid: %s (database: %s)", filename, database)
	}

	return filename, nil
}
