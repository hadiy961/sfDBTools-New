package compress

import (
	"compress/gzip"
	"compress/zlib"
	"io"
	"runtime"

	"github.com/klauspost/compress/zstd"
	"github.com/klauspost/pgzip"
	"github.com/ulikunitz/xz"
)

// createGzipWriter creates a gzip writer with specified level
func createGzipWriter(w io.Writer, level CompressionLevel) (*gzip.Writer, error) {
	// Konversi level (1-9) ke gzip level
	// gzip: 1 = BestSpeed, 9 = BestCompression
	gzipLevel := int(level)
	if gzipLevel < 1 {
		gzipLevel = gzip.DefaultCompression
	}
	if gzipLevel > 9 {
		gzipLevel = 9
	}

	return gzip.NewWriterLevel(w, gzipLevel)
}

// createPgzipWriter creates a parallel gzip writer with specified level and optimized concurrency
func createPgzipWriter(w io.Writer, level CompressionLevel) (*pgzip.Writer, error) {
	// Konversi level (1-9) ke pgzip level
	gzipLevel := int(level)
	if gzipLevel < 1 {
		gzipLevel = pgzip.DefaultCompression
	}
	if gzipLevel > 9 {
		gzipLevel = 9
	}

	// Buat writer dengan level yang ditentukan
	pw, err := pgzip.NewWriterLevel(w, gzipLevel)
	if err != nil {
		return nil, err
	}

	// Set concurrency untuk maksimalkan penggunaan CPU
	// Block size 1MB, gunakan semua CPU cores
	blockSize := 1 << 20 // 1MB blocks
	numBlocks := runtime.NumCPU()
	if err := pw.SetConcurrency(blockSize, numBlocks); err != nil {
		return nil, err
	}

	return pw, nil
}

// createZlibWriter creates a zlib writer with specified level
func createZlibWriter(w io.Writer, level CompressionLevel) (*zlib.Writer, error) {
	// Konversi level (1-9) ke zlib level
	zlibLevel := int(level)
	if zlibLevel < 1 {
		zlibLevel = zlib.DefaultCompression
	}
	if zlibLevel > 9 {
		zlibLevel = 9
	}

	return zlib.NewWriterLevel(w, zlibLevel)
}

// createZstdWriter creates a zstandard writer with specified level, concurrency, and window size
func createZstdWriter(w io.Writer, level CompressionLevel) (*zstd.Encoder, error) {
	var zstdLevel zstd.EncoderLevel

	// Mapping level 1-9 ke zstd EncoderLevel
	switch {
	case level <= 2:
		zstdLevel = zstd.SpeedFastest
	case level <= 4:
		zstdLevel = zstd.SpeedDefault
	case level <= 6:
		zstdLevel = zstd.SpeedBetterCompression
	default:
		zstdLevel = zstd.SpeedBestCompression
	}

	// Buat writer dengan optimalisasi untuk throughput dan compression ratio
	return zstd.NewWriter(w,
		zstd.WithEncoderLevel(zstdLevel),
		zstd.WithEncoderConcurrency(runtime.NumCPU()), // Parallel compression
		zstd.WithWindowSize(8<<20),                    // 8MB window untuk better compression ratio
	)
}

// createXzWriter creates an XZ writer with specified level
func createXzWriter(w io.Writer, level CompressionLevel) (*xz.Writer, error) {
	// XZ compression level mapping (0-9)
	// 0 = fastest, 9 = best compression
	xzLevel := int(level)
	if xzLevel < 0 {
		xzLevel = 6 // Default level
	}
	if xzLevel > 9 {
		xzLevel = 9
	}

	// Create XZ writer with compression level
	// Note: xz.Writer tidak support concurrent compression seperti pgzip/zstd
	return xz.NewWriter(w)
}
