package verify

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/cespare/xxhash/v2"
)

// GetHasher mengembalikan instance hasher berdasarkan algoritma yang dipilih
func GetHasher(algo string) (hash.Hash, error) {
	switch algo {
	case "md5":
		return md5.New(), nil
	case "sha256", "": // default
		return sha256.New(), nil
	case "xxhash":
		return xxhash.New(), nil
	default:
		return nil, fmt.Errorf("unsupported checksum algorithm: %s", algo)
	}
}

// HashToString mengembalikan string representasi (hex) dari hasher yang sudah di-sum
func HashToString(hasher hash.Hash) string {
	return hex.EncodeToString(hasher.Sum(nil))
}

// GenerateChecksum menghitung hash dari file on-disk (tanpa decrypt/decompress)
// Menggunakan io.CopyBuffer dari file ke hash.Hash untuk memory-efficient streaming
func GenerateChecksum(filePath string, algo string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer file.Close()

	hasher, err := GetHasher(algo)
	if err != nil {
		return "", err
	}

	buf := make([]byte, 256*1024) // 256KB read buffer
	if _, err := io.CopyBuffer(hasher, file, buf); err != nil {
		return "", fmt.Errorf("failed to read file for checksum: %w", err)
	}

	return HashToString(hasher), nil
}

// CompareChecksum menghitung hash dan membandingkan dengan expected
func CompareChecksum(filePath string, algo string, expected string) (match bool, actual string, err error) {
	actual, err = GenerateChecksum(filePath, algo)
	if err != nil {
		return false, "", err
	}
	return actual == expected, actual, nil
}
