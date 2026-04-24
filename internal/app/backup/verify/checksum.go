package verify

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

// GenerateChecksum menghitung hash dari file on-disk (tanpa decrypt/decompress)
// Menggunakan io.CopyBuffer dari file ke hash.Hash untuk memory-efficient streaming
func GenerateChecksum(filePath string, algo string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer file.Close()

	var hasher hash.Hash
	switch algo {
	case "md5":
		hasher = md5.New()
	case "sha256", "": // default to sha256
		hasher = sha256.New()
	default:
		return "", fmt.Errorf("unsupported checksum algorithm: %s", algo)
	}

	buf := make([]byte, 256*1024) // 256KB read buffer
	if _, err := io.CopyBuffer(hasher, file, buf); err != nil {
		return "", fmt.Errorf("failed to read file for checksum: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// CompareChecksum menghitung hash dan membandingkan dengan expected
func CompareChecksum(filePath string, algo string, expected string) (match bool, actual string, err error) {
	actual, err = GenerateChecksum(filePath, algo)
	if err != nil {
		return false, "", err
	}
	return actual == expected, actual, nil
}
