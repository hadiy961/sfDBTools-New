// File : internal/crypto/stream/writer.go
// Deskripsi : Streaming encryption writer dengan chunking untuk large files
// Author : Hadiyatna Muflihun
// Tanggal : 8 Januari 2026
// Last Modified : 8 Januari 2026
package stream

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"sfdbtools/internal/crypto/core"

	"golang.org/x/crypto/pbkdf2"
)

// Writer implements io.WriteCloser for streaming AES-256-GCM encryption.
//
// Data is encrypted in chunks to avoid loading entire files into memory.
// Each chunk is independently encrypted with a unique nonce derived from
// a base nonce + counter.
//
// Format:
//   - Header: "Salted__" (8 bytes)
//   - Salt: 8 bytes
//   - Base Nonce: 12 bytes
//   - Chunks: [4-byte size][encrypted data]...
//   - End marker: 4 zero bytes
//
// This is compatible with the corresponding Reader for decryption.
type Writer struct {
	writer       io.Writer
	gcm          cipher.AEAD
	baseNonce    []byte
	buffer       *bytes.Buffer
	chunkCounter uint64
	closed       bool
}

// NewWriter creates a new streaming encryption writer.
//
// Parameters:
//   - w: underlying writer (typically a file)
//   - passphrase: encryption password/key
//
// Returns:
//   - *Writer implementing io.WriteCloser
//   - error if initialization fails
//
// Important: Must call Close() to flush final chunk and write end marker.
func NewWriter(w io.Writer, passphrase []byte) (*Writer, error) {
	// Generate random salt
	salt := make([]byte, core.SaltSizeBytes)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key from passphrase and salt
	key := pbkdf2.Key(passphrase, salt, core.PBKDF2Iterations, 32, sha256.New)

	// Initialize AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random base nonce for chunking
	baseNonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, baseNonce); err != nil {
		return nil, fmt.Errorf("failed to generate base nonce: %w", err)
	}

	// Write header: "Salted__" + salt + base nonce
	opensslHeader := []byte(core.OpenSSLSaltedHeader)
	if _, err := w.Write(opensslHeader); err != nil {
		return nil, fmt.Errorf("failed to write header: %w", err)
	}
	if _, err := w.Write(salt); err != nil {
		return nil, fmt.Errorf("failed to write salt: %w", err)
	}
	if _, err := w.Write(baseNonce); err != nil {
		return nil, fmt.Errorf("failed to write base nonce: %w", err)
	}

	return &Writer{
		writer:       w,
		gcm:          gcm,
		baseNonce:    baseNonce,
		buffer:       bytes.NewBuffer(make([]byte, 0, core.StreamChunkSize)),
		chunkCounter: 0,
		closed:       false,
	}, nil
}

// Write encrypts data in chunks and writes to underlying writer.
//
// Data is buffered until a full chunk (64KB) is available, then encrypted
// and written. Partial chunks are kept in buffer until Close() is called.
//
// Returns number of bytes accepted (before encryption) and any error.
func (w *Writer) Write(p []byte) (n int, err error) {
	if w.closed {
		return 0, fmt.Errorf("writer already closed")
	}

	written := 0
	for len(p) > 0 {
		// Calculate how much can fit in current buffer
		available := core.StreamChunkSize - w.buffer.Len()
		toWrite := len(p)
		if toWrite > available {
			toWrite = available
		}

		// Write to buffer
		w.buffer.Write(p[:toWrite])
		written += toWrite
		p = p[toWrite:]

		// Flush chunk if buffer is full
		if w.buffer.Len() >= core.StreamChunkSize {
			if err := w.flushChunk(); err != nil {
				return written, fmt.Errorf("failed to flush chunk: %w", err)
			}
		}
	}

	return written, nil
}

// flushChunk encrypts and writes one chunk from buffer.
func (w *Writer) flushChunk() error {
	if w.buffer.Len() == 0 {
		return nil
	}

	// Get plaintext from buffer
	plaintext := w.buffer.Bytes()

	// Generate unique nonce for this chunk: baseNonce XOR counter
	nonce := make([]byte, len(w.baseNonce))
	copy(nonce, w.baseNonce)
	binary.BigEndian.PutUint64(nonce[len(nonce)-8:], w.chunkCounter)

	// Encrypt chunk
	ciphertext := w.gcm.Seal(nil, nonce, plaintext, nil)

	// Write chunk: [4-byte size][encrypted data]
	chunkSizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(chunkSizeBytes, uint32(len(ciphertext)))

	if _, err := w.writer.Write(chunkSizeBytes); err != nil {
		return fmt.Errorf("failed to write chunk size: %w", err)
	}
	if _, err := w.writer.Write(ciphertext); err != nil {
		return fmt.Errorf("failed to write ciphertext: %w", err)
	}

	// Reset buffer and increment counter
	w.buffer.Reset()
	w.chunkCounter++

	return nil
}

// Close flushes any remaining data and writes end marker.
//
// Must be called to properly finalize the encrypted stream.
// After Close(), no more Write() calls are allowed.
func (w *Writer) Close() error {
	if w.closed {
		return nil
	}

	// Flush any remaining data in buffer
	if w.buffer.Len() > 0 {
		if err := w.flushChunk(); err != nil {
			return fmt.Errorf("failed to flush final chunk: %w", err)
		}
	}

	// Write end marker (chunk size = 0)
	endMarker := make([]byte, 4)
	binary.BigEndian.PutUint32(endMarker, 0)
	if _, err := w.writer.Write(endMarker); err != nil {
		return fmt.Errorf("failed to write end marker: %w", err)
	}

	w.closed = true

	// Close underlying writer if it implements io.Closer
	if closer, ok := w.writer.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}
