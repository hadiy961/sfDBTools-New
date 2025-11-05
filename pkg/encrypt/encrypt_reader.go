// File : pkg/encrypt/encrypt_reader.go
// Deskripsi : Streaming decrypt reader untuk backup files
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
)

// DecryptingReader implements io.Reader untuk streaming decryption
type DecryptingReader struct {
	source     io.Reader
	passphrase string
	gcm        cipher.AEAD
	buffer     *bytes.Buffer
	nonce      []byte
	headerRead bool
	saltRead   bool
	nonceRead  bool
}

// NewDecryptingReader membuat reader untuk decrypt streaming data
func NewDecryptingReader(source io.Reader, passphrase string) (io.Reader, error) {
	return &DecryptingReader{
		source:     source,
		passphrase: passphrase,
		buffer:     new(bytes.Buffer),
		headerRead: false,
		saltRead:   false,
		nonceRead:  false,
	}, nil
}

// Read implements io.Reader interface
func (dr *DecryptingReader) Read(p []byte) (n int, err error) {
	// Untuk simplicity, kita baca seluruh file sekaligus dan decrypt
	// Ini lebih sederhana dibanding streaming decryption untuk AES-GCM
	if !dr.headerRead {
		// Baca seluruh encrypted data
		allData, err := io.ReadAll(dr.source)
		if err != nil {
			return 0, err
		}

		// Validate minimum length
		if len(allData) < 16 {
			return 0, fmt.Errorf("encrypted data too short")
		}

		// Check header "Salted__"
		opensslHeader := []byte("Salted__")
		if !bytes.Equal(allData[:8], opensslHeader) {
			return 0, fmt.Errorf("invalid encrypted format: missing 'Salted__' header")
		}

		// Extract salt
		salt := allData[8:16]

		// Extract ciphertext (contains nonce + encrypted data + auth tag)
		ciphertextWithNonce := allData[16:]

		// Derive key from passphrase and salt
		key := deriveKey([]byte(dr.passphrase), salt)

		// Initialize AES-GCM
		block, err := aes.NewCipher(key)
		if err != nil {
			return 0, err
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return 0, err
		}

		// Extract nonce (first NonceSize bytes of ciphertext)
		nonceSize := gcm.NonceSize()
		if len(ciphertextWithNonce) < nonceSize {
			return 0, fmt.Errorf("ciphertext too short")
		}

		nonce := ciphertextWithNonce[:nonceSize]
		ciphertext := ciphertextWithNonce[nonceSize:]

		// Decrypt
		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return 0, fmt.Errorf("decryption failed: %w", err)
		}

		// Write plaintext to buffer
		dr.buffer.Write(plaintext)
		dr.headerRead = true
	}

	// Read from buffer
	return dr.buffer.Read(p)
}
