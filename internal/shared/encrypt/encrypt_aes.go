// File : pkg/encrypt/encrypt_aes.go
// Deskripsi : Fungsi utilitas untuk enkripsi dan dekripsi AES kompatibel dengan openssl enc -pbkdf2
// Author : Hadiyatna Muflihun
// Tanggal : 3 Oktober 2024
// Last Modified : 3 Oktober 2024
package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"sfdbtools/internal/shared/consts"

	"golang.org/x/crypto/pbkdf2"
)

// deriveKey menggunakan PBKDF2 untuk membuat kunci dari password.
func deriveKey(passphrase, salt []byte) []byte {
	return pbkdf2.Key(passphrase, salt, consts.PBKDF2Iterations, 32, sha256.New) // 32 byte untuk AES-256
}

// EncryptAES mengenkripsi data agar kompatibel dengan `openssl enc -pbkdf2`.
func EncryptAES(plaintext []byte, passphrase []byte) ([]byte, error) {
	// Hasil enkripsi akan dalam format:
	// "Salted__" (8 byte) + salt (8 byte) + ciphertext (yang sudah mengandung nonce)
	salt := make([]byte, consts.SaltSizeBytes)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	// Turunkan kunci dari password dan salt
	key := deriveKey(passphrase, salt)

	// Inisialisasi AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Buat GCM (Galois/Counter Mode)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Buat nonce acak untuk GCM
	nonce := make([]byte, gcm.NonceSize())
	// Isi nonce dengan data acak
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// gcm.Seal akan menggabungkan nonce, data terenkripsi, dan tag otentikasi.
	// Kita akan menyimpannya langsung setelah salt.
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Format OpenSSL: "Salted__" (8 byte) + salt (8 byte) + ciphertext (yang sudah mengandung nonce)
	opensslHeader := []byte(consts.OpenSSLSaltedHeader)
	// Gabungkan semuanya
	encryptedPayload := append(opensslHeader, salt...)
	encryptedPayload = append(encryptedPayload, ciphertext...)

	// Kembalikan hasil enkripsi
	return encryptedPayload, nil
}

// DecryptAES mendekripsi data yang dienkripsi oleh `openssl enc -pbkdf2`.
func DecryptAES(encryptedPayload []byte, passphrase []byte) ([]byte, error) {
	// Cek header "Salted__" dan ekstrak salt
	opensslHeader := []byte(consts.OpenSSLSaltedHeader)

	// Validasi panjang payload
	if len(encryptedPayload) < 16 {
		return nil, fmt.Errorf("payload terenkripsi tidak valid: terlalu pendek")
	}

	// Cek header
	header := encryptedPayload[:8]
	if !bytes.Equal(header, opensslHeader) {
		return nil, fmt.Errorf("format file tidak valid: header 'Salted__' tidak ditemukan")
	}

	// Ekstrak salt dan ciphertext
	salt := encryptedPayload[8:16]
	ciphertextWithNonce := encryptedPayload[16:] // Sisa payload adalah ciphertext + nonce

	// Turunkan kunci dari password dan salt
	key := deriveKey(passphrase, salt)

	// Inisialisasi AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Buat GCM (Galois/Counter Mode)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Ekstrak nonce dari awal ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertextWithNonce) < nonceSize {
		return nil, fmt.Errorf("ciphertext tidak valid: lebih pendek dari ukuran nonce")
	}

	// Ekstrak nonce dan ciphertext asli
	nonce, ciphertext := ciphertextWithNonce[:nonceSize], ciphertextWithNonce[nonceSize:]

	// Dekripsi data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	// Kembalikan plaintext
	return plaintext, nil
}
