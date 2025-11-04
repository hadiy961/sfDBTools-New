package compress

import (
	"fmt"
	"strings"
)

// ValidateCompressionType validates if the compression type is supported
func ValidateCompressionType(compressionType string) (CompressionType, error) {
	ct := CompressionType(strings.ToLower(compressionType))
	switch ct {
	case CompressionNone, CompressionGzip, CompressionPgzip, CompressionZlib, CompressionZstd:
		return ct, nil
	default:
		return CompressionNone, fmt.Errorf("unsupported compression type: %s. Supported types: none, gzip, pgzip, zlib, zstd", compressionType)
	}
}

// ValidateCompressionLevel validates if the compression level is valid (1-9)
func ValidateCompressionLevel(level int) (CompressionLevel, error) {
	if level < 1 || level > 9 {
		return LevelDefault, fmt.Errorf("unsupported compression level: %d. Supported levels: 1-9 (1=fastest, 9=best)", level)
	}
	return CompressionLevel(level), nil
}
