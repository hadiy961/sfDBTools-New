package compress

import (
	"compress/zlib"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/klauspost/pgzip"
	"github.com/ulikunitz/xz"
)

// NewDecompressingReader returns a reader that decompresses data from r using the specified compression type.
// For zstd and pgzip, uses parallel decompression with all available CPU cores for optimal performance.
func NewDecompressingReader(r io.Reader, ctype CompressionType) (io.ReadCloser, error) {
	switch ctype {
	case CompressionGzip, CompressionPgzip:
		// Gunakan pgzip untuk semua .gz agar konsisten (ekstensi sama)
		pr, err := pgzip.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("gagal create pgzip reader: %w", err)
		}
		return pr, nil
	case CompressionZlib:
		return zlib.NewReader(r)
	case CompressionZstd:
		// Buat zstd decoder dengan concurrency untuk parallel decompression
		zr, err := zstd.NewReader(r,
			zstd.WithDecoderConcurrency(runtime.NumCPU()), // Parallel decoding
			zstd.WithDecoderLowmem(false),                 // Use more memory for speed
		)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(zr), nil
	case CompressionXz:
		// XZ decompression support
		xzReader, err := xz.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("gagal create xz reader: %w", err)
		}
		return io.NopCloser(xzReader), nil
	case CompressionNone:
		return io.NopCloser(r), nil
	default:
		return nil, fmt.Errorf("unsupported compression type: %s", ctype)
	}
}

// DetectCompressionTypeFromFile detects compression type based on file extension.
// Returns CompressionType for type safety.
func DetectCompressionTypeFromFile(path string) CompressionType {
	name := strings.ToLower(path)
	// Remove .enc extension first if exists
	name = strings.TrimSuffix(name, ".enc")

	switch {
	case strings.HasSuffix(name, ".gz"):
		// Selalu gunakan pgzip untuk ekstensi .gz
		return CompressionPgzip
	case strings.HasSuffix(name, ".zst"):
		return CompressionZstd
	case strings.HasSuffix(name, ".xz"):
		return CompressionXz
	case strings.HasSuffix(name, ".zlib"):
		return CompressionZlib
	default:
		return CompressionNone
	}
}
