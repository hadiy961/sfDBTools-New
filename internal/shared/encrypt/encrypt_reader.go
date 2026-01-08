// File : pkg/encrypt/encrypt_reader.go
// Deskripsi : Streaming decrypt reader untuk backup files
// Author : Hadiyatna Muflihun
// Tanggal : 5 November 2025
// Last Modified : 14 November 2025

package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"io"
)

// DecryptingReader implements io.Reader untuk streaming decryption dengan chunking
type DecryptingReader struct {
	source       io.Reader
	passphrase   string
	gcm          cipher.AEAD
	buffer       *bytes.Buffer
	baseNonce    []byte
	headerRead   bool
	chunkCounter uint64
	eof          bool
}

// NewDecryptingReader membuat reader untuk decrypt streaming data dengan chunking
func NewDecryptingReader(source io.Reader, passphrase string) (io.Reader, error) {
	return &DecryptingReader{
		source:       source,
		passphrase:   passphrase,
		buffer:       new(bytes.Buffer),
		headerRead:   false,
		chunkCounter: 0,
		eof:          false,
	}, nil
}

// Read implements io.Reader interface dengan chunked decryption
func (dr *DecryptingReader) Read(p []byte) (n int, err error) {
	// Jika sudah EOF, kembalikan dari buffer atau EOF
	if dr.eof && dr.buffer.Len() == 0 {
		return 0, io.EOF
	}

	// Baca header jika belum
	if !dr.headerRead {
		if err := dr.readHeader(); err != nil {
			return 0, err
		}
	}

	// Loop sampai buffer terisi atau EOF
	for dr.buffer.Len() < len(p) && !dr.eof {
		if err := dr.readAndDecryptChunk(); err != nil {
			if err == io.EOF {
				dr.eof = true
				break
			}
			return 0, err
		}
	}

	// Baca dari buffer
	return dr.buffer.Read(p)
}

// readHeader membaca header, salt, dan base nonce
func (dr *DecryptingReader) readHeader() error {
	// Baca header "Salted__" (8 bytes)
	opensslHeader := make([]byte, 8)
	if _, err := io.ReadFull(dr.source, opensslHeader); err != nil {
		return fmt.Errorf("gagal membaca header: %w", err)
	}
	if !bytes.Equal(opensslHeader, []byte("Salted__")) {
		return fmt.Errorf("invalid encrypted format: missing 'Salted__' header")
	}

	// Baca salt (8 bytes)
	salt := make([]byte, 8)
	if _, err := io.ReadFull(dr.source, salt); err != nil {
		return fmt.Errorf("gagal membaca salt: %w", err)
	}

	// Derive key dari passphrase dan salt
	key := deriveKey([]byte(dr.passphrase), salt)

	// Inisialisasi AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("gagal membuat cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("gagal membuat GCM: %w", err)
	}
	dr.gcm = gcm

	// Baca base nonce
	baseNonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(dr.source, baseNonce); err != nil {
		return fmt.Errorf("gagal membaca base nonce: %w", err)
	}
	dr.baseNonce = baseNonce

	dr.headerRead = true
	return nil
}

// readAndDecryptChunk membaca dan decrypt satu chunk
func (dr *DecryptingReader) readAndDecryptChunk() error {
	// Baca chunk size (4 bytes)
	chunkSizeBytes := make([]byte, 4)
	if _, err := io.ReadFull(dr.source, chunkSizeBytes); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return io.EOF
		}
		return fmt.Errorf("gagal membaca chunk size: %w", err)
	}

	chunkSize := binary.BigEndian.Uint32(chunkSizeBytes)

	// Jika chunk size = 0, ini adalah end marker
	if chunkSize == 0 {
		return io.EOF
	}

	// Baca ciphertext chunk
	ciphertext := make([]byte, chunkSize)
	if _, err := io.ReadFull(dr.source, ciphertext); err != nil {
		return fmt.Errorf("gagal membaca ciphertext chunk: %w", err)
	}

	// Generate nonce untuk chunk ini
	nonce := make([]byte, len(dr.baseNonce))
	copy(nonce, dr.baseNonce)
	binary.BigEndian.PutUint64(nonce[len(nonce)-8:], dr.chunkCounter)

	// Decrypt chunk
	plaintext, err := dr.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("gagal decrypt chunk %d: %w", dr.chunkCounter, err)
	}

	// Tulis plaintext ke buffer
	dr.buffer.Write(plaintext)
	dr.chunkCounter++

	return nil
}
