package compress

import (
	"compress/zlib"
	"fmt"
	"io"
	"runtime"
	"sfdbtools/internal/shared/consts"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/klauspost/pgzip"
	"github.com/ulikunitz/xz"
)

// NewDecompressingReader returns a reader that decompresses data from r using the specified compression type.
// For zstd and pgzip, uses parallel decompression with all available CPU cores for optimal performance.
func NewDecompressingReader(r io.Reader, ctype CompressionType) (io.ReadCloser, error) {
	switch ctype {
	case CompressionType(consts.CompressionTypeGzip), CompressionType(consts.CompressionTypePgzip):
		// Gunakan pgzip untuk semua .gz agar konsisten (ekstensi sama)
		pr, err := pgzip.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("gagal create pgzip reader: %w", err)
		}
		return pr, nil
	case CompressionType(consts.CompressionTypeZlib):
		return zlib.NewReader(r)
	case CompressionType(consts.CompressionTypeZstd):
		// Buat zstd decoder dengan concurrency untuk parallel decompression
		zr, err := zstd.NewReader(r,
			zstd.WithDecoderConcurrency(runtime.NumCPU()), // Parallel decoding
			zstd.WithDecoderLowmem(false),                 // Use more memory for speed
		)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(zr), nil
	case CompressionType(consts.CompressionTypeXz):
		// XZ decompression support
		xzReader, err := xz.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("gagal create xz reader: %w", err)
		}
		return io.NopCloser(xzReader), nil
	case CompressionType(consts.CompressionTypeNone):
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
	name = strings.TrimSuffix(name, consts.ExtEnc)

	switch {
	case strings.HasSuffix(name, consts.ExtGzip):
		// Selalu gunakan pgzip untuk ekstensi .gz
		return CompressionType(consts.CompressionTypePgzip)
	case strings.HasSuffix(name, consts.ExtZstd):
		return CompressionType(consts.CompressionTypeZstd)
	case strings.HasSuffix(name, consts.ExtXz):
		return CompressionType(consts.CompressionTypeXz)
	case strings.HasSuffix(name, consts.ExtZlib):
		return CompressionType(consts.CompressionTypeZlib)
	default:
		return CompressionType(consts.CompressionTypeNone)
	}
}
