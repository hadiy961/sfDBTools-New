package verify

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseMinFileSize mengkonversi string size (e.g. "1KB", "100B", "5MB") ke bytes
func ParseMinFileSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))
	if sizeStr == "" {
		return 0, nil
	}

	var multiplier int64 = 1
	var numStr string

	if strings.HasSuffix(sizeStr, "KB") {
		multiplier = 1024
		numStr = strings.TrimSuffix(sizeStr, "KB")
	} else if strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1024 * 1024
		numStr = strings.TrimSuffix(sizeStr, "MB")
	} else if strings.HasSuffix(sizeStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(sizeStr, "GB")
	} else if strings.HasSuffix(sizeStr, "B") {
		multiplier = 1
		numStr = strings.TrimSuffix(sizeStr, "B")
	} else {
		// assume bytes if no suffix
		numStr = sizeStr
	}

	numStr = strings.TrimSpace(numStr)
	val, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size format '%s': %w", sizeStr, err)
	}

	return val * multiplier, nil
}

// ValidateSize mengecek apakah file size >= minimum threshold
func ValidateSize(filePath string, minSize int64) (bool, int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return false, 0, fmt.Errorf("failed to stat file: %w", err)
	}

	size := info.Size()
	return size >= minSize, size, nil
}
