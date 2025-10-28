package helper

import (
	"os"
	"strconv"
)

// Helper functions untuk mengambil environment variables

// GetEnvOrDefault mengambil nilai dari environment variable atau mengembalikan defaultValue jika tidak ada
func GetEnvOrDefault(key, defaultValue string) string {
	return os.Getenv(key)
}

// GetEnvOrDefaultInt mengambil nilai integer dari environment variable atau mengembalikan defaultValue jika tidak ada atau tidak valid
func GetEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
