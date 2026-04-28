package verify

import (
	"bufio"
	"fmt"
	"hash"
	"io"
	"os"
	backupfile "sfdbtools/internal/app/backup/helpers/file"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/shared/compress"
	"sfdbtools/internal/shared/consts"
)

// OpenVerifyReader membuka file dan menyiapkan reader dengan decrypt/decompress
// Jika hasher diberikan (!= nil), maka akan disisipkan io.TeeReader di layer paling bawah
// sehingga pembacaan stream akan langsung dikalkulasikan hash-nya.
func OpenVerifyReader(filePath string, encryptionKey string, hasher hash.Hash) (io.Reader, []io.Closer, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal membuka file: %w", err)
	}

	var baseReader io.Reader = file
	if hasher != nil {
		baseReader = io.TeeReader(file, hasher)
	}

	// Buffer file reads to improve large sequential throughput
	reader := io.Reader(bufio.NewReaderSize(baseReader, 4*1024*1024))
	closers := []io.Closer{file}

	// Decrypt if encrypted
	isEncrypted := backupfile.IsEncryptedFile(filePath)
	if isEncrypted {
		if encryptionKey == "" {
			CloseReaders(closers)
			return nil, nil, fmt.Errorf("file is encrypted but no encryption key provided")
		}
		decReader, err := crypto.NewStreamDecryptor(reader, encryptionKey)
		if err != nil {
			CloseReaders(closers)
			return nil, nil, fmt.Errorf("gagal membuat decrypting reader: %w", err)
		}
		reader = decReader
		closers = append(closers, io.NopCloser(decReader))
	}

	// Decompress if compressed
	compressionType := compress.DetectCompressionTypeFromFile(filePath)
	if compressionType != compress.CompressionType(consts.CompressionTypeNone) {
		decompReader, err := compress.NewDecompressingReader(reader, compressionType)
		if err != nil {
			CloseReaders(closers)
			return nil, nil, fmt.Errorf("gagal membuat decompressing reader: %w", err)
		}
		reader = decompReader
		closers = append(closers, decompReader)
	}

	return reader, closers, nil
}

// CloseReaders menutup semua readers dengan urutan terbalik
func CloseReaders(closers []io.Closer) {
	for i := len(closers) - 1; i >= 0; i-- {
		if closer := closers[i]; closer != nil {
			_ = closer.Close()
		}
	}
}
