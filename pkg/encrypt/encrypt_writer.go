// File : pkg/encrypt/encrypt_writer.go
// Deskripsi : Writer untuk enkripsi streaming AES
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-14
// Last Modified : 2025-11-14
package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// chunkSize adalah ukuran chunk untuk enkripsi streaming (64KB)
	chunkSize = 64 * 1024
)

// EncryptingWriter adalah writer yang mengenkripsi data secara streaming dengan chunking
type EncryptingWriter struct {
	writer        io.Writer
	gcm           cipher.AEAD
	baseNonce     []byte
	buffer        *bytes.Buffer
	headerWritten bool
	closed        bool
	chunkCounter  uint64
}

// NewEncryptingWriter membuat writer baru untuk enkripsi streaming dengan chunking
func NewEncryptingWriter(writer io.Writer, passphrase []byte) (*EncryptingWriter, error) {
	// Generate salt acak
	salt := make([]byte, saltSizeBytes)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("gagal generate salt: %w", err)
	}

	// Derive key dari passphrase dan salt
	key := pbkdf2.Key(passphrase, salt, pbkdf2Iterations, 32, sha256.New)

	// Inisialisasi AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat GCM: %w", err)
	}

	// Generate base nonce acak untuk chunking
	baseNonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, baseNonce); err != nil {
		return nil, fmt.Errorf("gagal generate base nonce: %w", err)
	}

	// Tulis header OpenSSL dan salt
	opensslHeader := []byte("Salted__")
	if _, err := writer.Write(opensslHeader); err != nil {
		return nil, fmt.Errorf("gagal menulis header: %w", err)
	}
	if _, err := writer.Write(salt); err != nil {
		return nil, fmt.Errorf("gagal menulis salt: %w", err)
	}

	// Tulis base nonce sebagai bagian awal ciphertext
	if _, err := writer.Write(baseNonce); err != nil {
		return nil, fmt.Errorf("gagal menulis base nonce: %w", err)
	}

	return &EncryptingWriter{
		writer:        writer,
		gcm:           gcm,
		baseNonce:     baseNonce,
		buffer:        bytes.NewBuffer(make([]byte, 0, chunkSize)),
		headerWritten: true,
		chunkCounter:  0,
	}, nil
}

// Write mengenkripsi data secara streaming dengan chunking
func (ew *EncryptingWriter) Write(p []byte) (n int, err error) {
	if ew.closed {
		return 0, fmt.Errorf("writer sudah ditutup")
	}

	// Tambahkan data ke buffer
	written := 0
	for len(p) > 0 {
		// Hitung berapa banyak yang bisa ditulis ke buffer
		available := chunkSize - ew.buffer.Len()
		toWrite := len(p)
		if toWrite > available {
			toWrite = available
		}

		// Tulis ke buffer
		ew.buffer.Write(p[:toWrite])
		written += toWrite
		p = p[toWrite:]

		// Jika buffer penuh, flush chunk
		if ew.buffer.Len() >= chunkSize {
			if err := ew.flushChunk(); err != nil {
				return written, err
			}
		}
	}

	return written, nil
}

// flushChunk mengenkripsi dan menulis satu chunk dari buffer
func (ew *EncryptingWriter) flushChunk() error {
	if ew.buffer.Len() == 0 {
		return nil
	}

	// Ambil data dari buffer
	plaintext := ew.buffer.Bytes()

	// Generate nonce unik untuk chunk ini dengan menggabungkan baseNonce + counter
	nonce := make([]byte, len(ew.baseNonce))
	copy(nonce, ew.baseNonce)
	// XOR counter ke dalam nonce untuk uniqueness
	binary.BigEndian.PutUint64(nonce[len(nonce)-8:], ew.chunkCounter)

	// Enkripsi chunk
	ciphertext := ew.gcm.Seal(nil, nonce, plaintext, nil)

	// Tulis ukuran chunk (4 bytes) + ciphertext
	chunkSizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(chunkSizeBytes, uint32(len(ciphertext)))
	if _, err := ew.writer.Write(chunkSizeBytes); err != nil {
		return fmt.Errorf("gagal menulis chunk size: %w", err)
	}
	if _, err := ew.writer.Write(ciphertext); err != nil {
		return fmt.Errorf("gagal menulis ciphertext: %w", err)
	}

	// Reset buffer dan increment counter
	ew.buffer.Reset()
	ew.chunkCounter++

	return nil
}

// Close menutup writer dan menulis chunk terakhir jika ada
func (ew *EncryptingWriter) Close() error {
	if ew.closed {
		return nil
	}

	// Flush chunk terakhir jika buffer tidak kosong
	if ew.buffer.Len() > 0 {
		if err := ew.flushChunk(); err != nil {
			return err
		}
	}

	// Tulis marker akhir file (chunk size = 0)
	endMarker := make([]byte, 4)
	binary.BigEndian.PutUint32(endMarker, 0)
	if _, err := ew.writer.Write(endMarker); err != nil {
		return fmt.Errorf("gagal menulis end marker: %w", err)
	}

	ew.closed = true

	// Jika underlying writer memiliki Close method, panggil
	if closer, ok := ew.writer.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}
