package compress

import "io"

// CompressionType represents the type of compression algorithm
type CompressionType string

// CompressionLevel represents the compression level (1-9)
// 1 = fastest/least compression, 9 = slowest/best compression
type CompressionLevel int

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
