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

// normalizeLevel menormalisasi compression level ke range 1-9.
func normalizeLevel(level CompressionLevel, defaultLevel int) int {
	l := int(level)
	if l < 1 {
		return defaultLevel
	}
	if l > 9 {
		return 9
	}
	return l
}

func createGzipWriter(w io.Writer, level CompressionLevel) (*gzip.Writer, error) {
	return gzip.NewWriterLevel(w, normalizeLevel(level, gzip.DefaultCompression))
}

func createPgzipWriter(w io.Writer, level CompressionLevel) (*pgzip.Writer, error) {
	pw, err := pgzip.NewWriterLevel(w, normalizeLevel(level, pgzip.DefaultCompression))
	if err != nil {
		return nil, err
	}
	return pw, pw.SetConcurrency(1<<20, runtime.NumCPU())
}

func createZlibWriter(w io.Writer, level CompressionLevel) (*zlib.Writer, error) {
	return zlib.NewWriterLevel(w, normalizeLevel(level, zlib.DefaultCompression))
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

func createXzWriter(w io.Writer, level CompressionLevel) (*xz.Writer, error) {
	return xz.NewWriter(w)
}
