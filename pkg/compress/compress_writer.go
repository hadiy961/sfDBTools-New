package compress

import (
	"compress/gzip"
	"compress/zlib"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/klauspost/pgzip"
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

// createPgzipWriter creates a parallel gzip writer with specified level
func createPgzipWriter(w io.Writer, level CompressionLevel) (*pgzip.Writer, error) {
	// Konversi level (1-9) ke pgzip level
	gzipLevel := int(level)
	if gzipLevel < 1 {
		gzipLevel = pgzip.DefaultCompression
	}
	if gzipLevel > 9 {
		gzipLevel = 9
	}

	return pgzip.NewWriterLevel(w, gzipLevel)
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

// createZstdWriter creates a zstandard writer with specified level
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

	return zstd.NewWriter(w, zstd.WithEncoderLevel(zstdLevel))
}
