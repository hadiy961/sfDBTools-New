package file

import (
	"path/filepath"
	"strings"

	"sfDBTools/pkg/consts"
)

// ExtractDatabaseNameFromFile mengekstrak nama database dari nama file backup.
func ExtractDatabaseNameFromFile(filePath string) string {
	if filePath == "" {
		return ""
	}

	fileName := filepath.Base(filePath)

	fileName = strings.TrimSuffix(fileName, consts.ExtEnc)
	fileName = strings.TrimSuffix(fileName, consts.ExtGzip)
	fileName = strings.TrimSuffix(fileName, consts.ExtXz)
	fileName = strings.TrimSuffix(fileName, consts.ExtZstd)
	fileName = strings.TrimSuffix(fileName, consts.ExtZlib)
	fileName = strings.TrimSuffix(fileName, consts.ExtSQL)

	parts := strings.Split(fileName, "_")
	if len(parts) == 0 {
		return fileName
	}

	cutIndex := len(parts)
	for i := 0; i < len(parts); i++ {
		part := parts[i]

		if len(part) == 8 && isAllDigits(part) {
			cutIndex = i
			break
		}

		if len(part) == 6 && isAllDigits(part) && i > 0 {
			if len(parts[i-1]) == 8 && isAllDigits(parts[i-1]) {
				cutIndex = i - 1
				break
			}
		}

		if part == "backup" || part == "dump" || part == "pre" || part == "restore" {
			cutIndex = i
			break
		}

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
