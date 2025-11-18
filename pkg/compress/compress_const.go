package compress

import "io"

// CompressionType represents the type of compression algorithm
type CompressionType string

const (
	CompressionNone  CompressionType = "none"
	CompressionGzip  CompressionType = "gzip"  // Standard gzip (single-threaded)
	CompressionPgzip CompressionType = "pgzip" // Parallel gzip (multi-threaded)
	CompressionZlib  CompressionType = "zlib"  // DEFLATE (zlib format)
	CompressionZstd  CompressionType = "zstd"  // Zstandard (fast & good ratio)
	CompressionXz    CompressionType = "xz"    // LZMA/XZ (best ratio, slower)
)

// CompressionLevel represents the compression level (1-9)
// 1 = fastest/least compression, 9 = slowest/best compression
type CompressionLevel int

const (
	LevelBestSpeed CompressionLevel = 1
	LevelFast      CompressionLevel = 3
	LevelDefault   CompressionLevel = 6
	LevelBetter    CompressionLevel = 7
	LevelBest      CompressionLevel = 9
)

// CompressionConfig holds compression configuration
type CompressionConfig struct {
	Type  CompressionType
	Level CompressionLevel
}

// CompressingWriter wraps an io.Writer with compression
type CompressingWriter struct {
	baseWriter      io.Writer
	compressor      io.WriteCloser
	compressionType CompressionType
}
