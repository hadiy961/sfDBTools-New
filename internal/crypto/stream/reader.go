// File : internal/crypto/stream/reader.go
// Deskripsi : Streaming decryption reader untuk encrypted files
// Author : Hadiyatna Muflihun
// Tanggal : 8 Januari 2026
// Last Modified : 8 Januari 2026
package stream

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"sfdbtools/internal/crypto/core"

	"golang.org/x/crypto/pbkdf2"
)

// Reader implements io.Reader for streaming AES-256-GCM decryption.
//
// Reads and decrypts chunked encrypted data produced by Writer.
// Data is decrypted chunk-by-chunk to avoid loading entire files into memory.
type Reader struct {
	source       io.Reader
	passphrase   string
	gcm          cipher.AEAD
	buffer       *bytes.Buffer
	baseNonce    []byte
	headerRead   bool
	chunkCounter uint64
	eof          bool
}

// NewReader creates a new streaming decryption reader.
//
// Parameters:
//   - source: underlying reader (typically an encrypted file)
//   - passphrase: decryption password/key
//
// Returns:
//   - io.Reader for reading decrypted data
//   - error if initialization fails
func NewReader(source io.Reader, passphrase string) (io.Reader, error) {
	return &Reader{
		source:       source,
		passphrase:   passphrase,
		buffer:       new(bytes.Buffer),
		headerRead:   false,
		chunkCounter: 0,
		eof:          false,
	}, nil
}

// Read implements io.Reader interface with chunked decryption.
//
// Reads encrypted chunks from source, decrypts them, and buffers the
// plaintext for consumption by the caller.
func (r *Reader) Read(p []byte) (n int, err error) {
	// If already EOF and buffer is empty, return EOF
	if r.eof && r.buffer.Len() == 0 {
		return 0, io.EOF
	}

	// Read header on first call
	if !r.headerRead {
		if err := r.readHeader(); err != nil {
			return 0, fmt.Errorf("failed to read header: %w", err)
		}
	}

	// Decrypt chunks until buffer has enough data or EOF
	for r.buffer.Len() < len(p) && !r.eof {
		if err := r.readAndDecryptChunk(); err != nil {
			if err == io.EOF {
				r.eof = true
				break
			}
			return 0, err
		}
	}

	// Read from buffer
	return r.buffer.Read(p)
}

// readHeader reads and validates header, salt, and base nonce.
func (r *Reader) readHeader() error {
	// Read "Salted__" header (8 bytes)
	opensslHeader := make([]byte, 8)
	if _, err := io.ReadFull(r.source, opensslHeader); err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}
	if !bytes.Equal(opensslHeader, []byte(core.OpenSSLSaltedHeader)) {
		return fmt.Errorf("invalid encrypted format: missing '%s' header", core.OpenSSLSaltedHeader)
	}

	// Read salt (8 bytes)
	salt := make([]byte, core.SaltSizeBytes)
	if _, err := io.ReadFull(r.source, salt); err != nil {
		return fmt.Errorf("failed to read salt: %w", err)
	}

	// Derive key from passphrase and salt
	key := pbkdf2.Key([]byte(r.passphrase), salt, core.PBKDF2Iterations, 32, sha256.New)

	// Initialize AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}
	r.gcm = gcm

	// Read base nonce
	baseNonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(r.source, baseNonce); err != nil {
		return fmt.Errorf("failed to read base nonce: %w", err)
	}
	r.baseNonce = baseNonce

	r.headerRead = true
	return nil
}

// readAndDecryptChunk reads and decrypts one chunk into buffer.
func (r *Reader) readAndDecryptChunk() error {
	// Read chunk size (4 bytes)
	chunkSizeBytes := make([]byte, 4)
	if _, err := io.ReadFull(r.source, chunkSizeBytes); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return io.EOF
		}
		return fmt.Errorf("failed to read chunk size: %w", err)
	}

	chunkSize := binary.BigEndian.Uint32(chunkSizeBytes)

	// Chunk size = 0 is end marker
	if chunkSize == 0 {
		return io.EOF
	}

	// Read encrypted chunk
	ciphertext := make([]byte, chunkSize)
	if _, err := io.ReadFull(r.source, ciphertext); err != nil {
		return fmt.Errorf("failed to read ciphertext chunk: %w", err)
	}

	// Generate nonce for this chunk: baseNonce XOR counter
	nonce := make([]byte, len(r.baseNonce))
	copy(nonce, r.baseNonce)
	binary.BigEndian.PutUint64(nonce[len(nonce)-8:], r.chunkCounter)

	// Decrypt chunk
	plaintext, err := r.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("failed to decrypt chunk %d (wrong key or corrupted data): %w",
			r.chunkCounter, err)
	}

	// Write plaintext to buffer
	r.buffer.Write(plaintext)
	r.chunkCounter++

	return nil
}
