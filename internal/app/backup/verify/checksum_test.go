package verify

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/cespare/xxhash/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetHasher(t *testing.T) {
	tests := []struct {
		algo        string
		expectError bool
	}{
		{"md5", false},
		{"sha256", false},
		{"", false}, // defaults to sha256
		{"xxhash", false},
		{"unsupported", true},
	}

	for _, tt := range tests {
		t.Run(tt.algo, func(t *testing.T) {
			hasher, err := GetHasher(tt.algo)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, hasher)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, hasher)
			}
		})
	}
}

func TestGenerateAndCompareChecksum(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_file.txt")
	content := []byte("hello world")
	err := os.WriteFile(filePath, content, 0644)
	assert.NoError(t, err)

	// Calculate expected hashes
	sha256Hasher := sha256.New()
	sha256Hasher.Write(content)
	expectedSha256 := hex.EncodeToString(sha256Hasher.Sum(nil))

	xxhasher := xxhash.New()
	xxhasher.Write(content)
	expectedXxhash := hex.EncodeToString(xxhasher.Sum(nil))

	md5Hasher := md5.New()
	md5Hasher.Write(content)
	expectedMd5 := hex.EncodeToString(md5Hasher.Sum(nil))

	tests := []struct {
		algo     string
		expected string
	}{
		{"sha256", expectedSha256},
		{"", expectedSha256},
		{"xxhash", expectedXxhash},
		{"md5", expectedMd5},
	}

	for _, tt := range tests {
		t.Run(tt.algo, func(t *testing.T) {
			// Test GenerateChecksum
			hash, err := GenerateChecksum(filePath, tt.algo)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, hash)

			// Test CompareChecksum
			match, actual, err := CompareChecksum(filePath, tt.algo, tt.expected)
			assert.NoError(t, err)
			assert.True(t, match)
			assert.Equal(t, tt.expected, actual)

			// Test CompareChecksum with wrong hash
			match, _, err = CompareChecksum(filePath, tt.algo, "wronghash")
			assert.NoError(t, err)
			assert.False(t, match)
		})
	}
}
