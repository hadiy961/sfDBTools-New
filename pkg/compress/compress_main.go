package compress

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	default:
		return ""
	}
}

// CompressFile compresses an existing file and creates a new compressed file
func CompressFile(inputPath, outputPath string, config CompressionConfig) error {
	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Create compressing writer
	compressingWriter, err := NewCompressingWriter(outputFile, config)
	if err != nil {
		return fmt.Errorf("failed to create compressing writer: %w", err)
	}
	defer compressingWriter.Close()

	// Copy data through compressor
	if _, err := io.Copy(compressingWriter, inputFile); err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}

	return nil
}

// GetCompressionInfo returns information about available compression types
func GetCompressionInfo() map[CompressionType]string {
	return map[CompressionType]string{
		CompressionNone:  "No compression",
		CompressionGzip:  "Standard gzip compression",
		CompressionPgzip: "Parallel gzip compression (faster for large files)",
		CompressionZlib:  "Zlib compression (good compression ratio)",
		CompressionZstd:  "Zstandard compression (fast and good ratio)",
	}
}
