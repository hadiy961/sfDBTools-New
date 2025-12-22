package env

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GetEnvOrDefault mengambil nilai dari environment variable atau mengembalikan defaultValue jika tidak ada.
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvOrDefaultInt mengambil nilai integer dari environment variable atau mengembalikan defaultValue jika tidak ada atau tidak valid.
func GetEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// ExpandPath mengexpand tilde (~) menjadi home directory.
func ExpandPath(path string) string {
	if path == "" {
		return path
	}

	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}

		if path == "~" {
			return home
		}

		return filepath.Join(home, path[2:])
	}

	return path
}
