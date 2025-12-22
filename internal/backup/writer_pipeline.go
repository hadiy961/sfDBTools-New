package backup

import (
	"fmt"
	"io"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/encrypt"
)

// createWriterPipeline membuat writer pipeline untuk compression dan encryption
func (s *Service) createWriterPipeline(baseWriter io.Writer, compressionRequired bool, compressionType string, encryptionKey string) (io.Writer, []io.Closer, error) {
	var writer io.Writer = baseWriter
	var closers []io.Closer

	// Layer 1: Encryption (paling dekat dengan file)
	if s.BackupDBOptions.Encryption.Enabled {
		encryptingWriter, err := encrypt.NewEncryptingWriter(writer, []byte(encryptionKey))
		if err != nil {
			return nil, nil, fmt.Errorf("gagal membuat encrypting writer: %w", err)
		}
		closers = append(closers, encryptingWriter)
		writer = encryptingWriter
	}

	// Layer 2: Compression
	if compressionRequired {
		compressionConfig := compress.CompressionConfig{
			Type:  compress.CompressionType(compressionType),
			Level: compress.CompressionLevel(s.BackupDBOptions.Compression.Level),
		}
		compressingWriter, err := compress.NewCompressingWriter(writer, compressionConfig)
		if err != nil {
			return nil, nil, fmt.Errorf("gagal membuat compressing writer: %w", err)
		}
		closers = append(closers, compressingWriter)
		writer = compressingWriter
	}

	return writer, closers, nil
}
