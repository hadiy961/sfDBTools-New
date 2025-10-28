// File : pkg/encrypt/encrypt_writer.go
// Deskripsi : Writer untuk enkripsi streaming AES
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-14
// Last Modified : 2024-10-14
package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// EncryptingWriter adalah writer yang mengenkripsi data secara streaming
type EncryptingWriter struct {
	writer        io.Writer
	gcm           cipher.AEAD
	nonce         []byte
	buffer        *bytes.Buffer
	headerWritten bool
	closed        bool
}

// NewEncryptingWriter membuat writer baru untuk enkripsi streaming
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

	// Generate nonce acak
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("gagal generate nonce: %w", err)
	}

	// Tulis header OpenSSL dan salt
	opensslHeader := []byte("Salted__")
	if _, err := writer.Write(opensslHeader); err != nil {
		return nil, fmt.Errorf("gagal menulis header: %w", err)
	}
	if _, err := writer.Write(salt); err != nil {
		return nil, fmt.Errorf("gagal menulis salt: %w", err)
	}

	// Tulis nonce sebagai bagian awal ciphertext
	if _, err := writer.Write(nonce); err != nil {
		return nil, fmt.Errorf("gagal menulis nonce: %w", err)
	}

	return &EncryptingWriter{
		writer:        writer,
		gcm:           gcm,
		nonce:         nonce,
		buffer:        bytes.NewBuffer(nil),
		headerWritten: true,
	}, nil
}

// Write mengenkripsi data dan menulisnya ke underlying writer
func (ew *EncryptingWriter) Write(p []byte) (n int, err error) {
	if ew.closed {
		return 0, fmt.Errorf("writer sudah ditutup")
	}

	// Tambahkan data ke buffer
	ew.buffer.Write(p)
	return len(p), nil
}

// Close menutup writer dan menulis data terenkripsi terakhir
func (ew *EncryptingWriter) Close() error {
	if ew.closed {
		return nil
	}

	// Enkripsi semua data yang ada di buffer
	plaintext := ew.buffer.Bytes()

	// Enkripsi data menggunakan GCM
	ciphertext := ew.gcm.Seal(nil, ew.nonce, plaintext, nil)

	// Tulis ciphertext ke underlying writer
	if _, err := ew.writer.Write(ciphertext); err != nil {
		return fmt.Errorf("gagal menulis ciphertext: %w", err)
	}

	ew.closed = true

	// Jika underlying writer memiliki Close method, panggil
	if closer, ok := ew.writer.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}
