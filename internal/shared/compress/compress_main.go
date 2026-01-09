package compress

import (
	"fmt"
	"io"
	"sfdbtools/internal/shared/consts"
)

// NewCompressingWriter creates a new compressing writer
func NewCompressingWriter(baseWriter io.Writer, config CompressionConfig) (*CompressingWriter, error) {
	if config.Type == CompressionType(consts.CompressionTypeNone) {
		return &CompressingWriter{
			baseWriter:      baseWriter,
			compressor:      nil,
			compressionType: CompressionType(consts.CompressionTypeNone),
		}, nil
	}

	var compressor io.WriteCloser
	var err error

	switch config.Type {
	case CompressionType(consts.CompressionTypeGzip):
		compressor, err = createGzipWriter(baseWriter, config.Level)
	case CompressionType(consts.CompressionTypePgzip):
		compressor, err = createPgzipWriter(baseWriter, config.Level)
	case CompressionType(consts.CompressionTypeZlib):
		compressor, err = createZlibWriter(baseWriter, config.Level)
	case CompressionType(consts.CompressionTypeZstd):
		compressor, err = createZstdWriter(baseWriter, config.Level)
	case CompressionType(consts.CompressionTypeXz):
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
	case CompressionType(consts.CompressionTypeGzip), CompressionType(consts.CompressionTypePgzip):
		return consts.ExtGzip
	case CompressionType(consts.CompressionTypeZlib):
		return consts.ExtZlib
	case CompressionType(consts.CompressionTypeZstd):
		return consts.ExtZstd
	case CompressionType(consts.CompressionTypeXz):
		return consts.ExtXz
	default:
		return ""
	}
}
