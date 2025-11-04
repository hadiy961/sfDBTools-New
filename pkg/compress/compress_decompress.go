package compress

import (
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"strings"

	"github.com/klauspost/compress/zstd"
)

// NewDecompressingReader returns a reader that decompresses data from r using the specified compression type.
func NewDecompressingReader(r io.Reader, ctype CompressionType) (io.ReadCloser, error) {
	switch ctype {
	case CompressionGzip, CompressionPgzip:
		return gzip.NewReader(r)
	case CompressionZlib:
		return zlib.NewReader(r)
	case CompressionZstd:
		zr, err := zstd.NewReader(r)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(zr), nil
	case CompressionNone:
		return io.NopCloser(r), nil
	default:
		return nil, fmt.Errorf("unsupported compression type: %s", ctype)
	}
}

// DetectCompressionTypeFromFile detects compression type based on file extension.
func DetectCompressionTypeFromFile(path string) CompressionType {
	name := strings.ToLower(path)
	name = strings.TrimSuffix(name, ".enc")
	switch {
	case strings.HasSuffix(name, ".gz"):
		return CompressionGzip
	case strings.HasSuffix(name, ".zst"):
		return CompressionZstd
	case strings.HasSuffix(name, ".zlib"):
		return CompressionZlib
	default:
		return CompressionNone
	}
}
