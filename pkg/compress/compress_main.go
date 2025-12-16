package compress

import (
	"fmt"
	"io"
)

// NewCompressingWriter creates a new compressing writer
func NewCompressingWriter(baseWriter io.Writer, config CompressionConfig) (*CompressingWriter, error) {
	if config.Type == CompressionNone {
		return &CompressingWriter{
			baseWriter:      baseWriter,
			compressor:      nil,
			compressionType: CompressionNone,
		}, nil
	}

	var compressor io.WriteCloser
	var err error

	switch config.Type {
	case CompressionGzip:
		compressor, err = createGzipWriter(baseWriter, config.Level)
	case CompressionPgzip:
		compressor, err = createPgzipWriter(baseWriter, config.Level)
	case CompressionZlib:
		compressor, err = createZlibWriter(baseWriter, config.Level)
	case CompressionZstd:
		compressor, err = createZstdWriter(baseWriter, config.Level)
	case CompressionXz:
		compressor, err = createXzWriter(baseWriter, config.Level)
	default:
		return nil, fmt.Errorf("unsupported compression type: %s", config.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create compressor: %w", err)
	}

	return &CompressingWriter{
		baseWriter:      baseWriter,
		compressor:      compressor,
		compressionType: config.Type,
	}, nil
}

// Write writes data through the compressor
func (cw *CompressingWriter) Write(p []byte) (n int, err error) {
	if cw.compressor == nil {
		return cw.baseWriter.Write(p)
	}
	return cw.compressor.Write(p)
}

// Close closes the compressor
func (cw *CompressingWriter) Close() error {
	if cw.compressor != nil {
		return cw.compressor.Close()
	}
	return nil
}

// GetFileExtension returns the appropriate file extension for the compression type
func GetFileExtension(compressionType CompressionType) string {
	switch compressionType {
	case CompressionGzip, CompressionPgzip:
		return ".gz"
	case CompressionZlib:
		return ".zlib"
	case CompressionZstd:
		return ".zst"
	case CompressionXz:
		return ".xz"
	default:
		return ""
	}
}
