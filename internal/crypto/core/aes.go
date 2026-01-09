// File : internal/crypto/core/aes.go
// Deskripsi : Core AES-256-GCM encryption/decryption primitives (OpenSSL compatible)
// Author : Hadiyatna Muflihun
// Tanggal : 8 Januari 2026
// Last Modified : 8 Januari 2026
package core

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

// deriveKey derives AES-256 key from passphrase using PBKDF2-SHA256.
// Uses 600,000 iterations per OWASP recommendation (2023).
func deriveKey(passphrase, salt []byte) []byte {
	return pbkdf2.Key(passphrase, salt, PBKDF2Iterations, 32, sha256.New)
}

// EncryptAES encrypts plaintext using AES-256-GCM with OpenSSL-compatible format.
//
// Format: "Salted__" (8 bytes) + salt (8 bytes) + nonce + ciphertext + auth tag
//
// This format is compatible with: openssl enc -aes-256-gcm -pbkdf2 -iter 600000
//
// Parameters:
//   - plaintext: data to encrypt
//   - passphrase: encryption password/key
//
// Returns:
//   - encrypted payload with OpenSSL header
//   - error if encryption fails
func EncryptAES(plaintext, passphrase []byte) ([]byte, error) {
	// Generate random salt
	salt := make([]byte, SaltSizeBytes)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key from passphrase and salt
	key := deriveKey(passphrase, salt)

	// Initialize AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM (Galois/Counter Mode) for authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data (Seal prepends nonce and appends auth tag)
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Build OpenSSL-compatible payload: header + salt + ciphertext
	opensslHeader := []byte(OpenSSLSaltedHeader)
	encryptedPayload := make([]byte, 0, len(opensslHeader)+len(salt)+len(ciphertext))
	encryptedPayload = append(encryptedPayload, opensslHeader...)
	encryptedPayload = append(encryptedPayload, salt...)
	encryptedPayload = append(encryptedPayload, ciphertext...)

	return encryptedPayload, nil
}

// DecryptAES decrypts OpenSSL-compatible AES-256-GCM encrypted data.
//
// Expected format: "Salted__" (8 bytes) + salt (8 bytes) + nonce + ciphertext + auth tag
//
// Parameters:
//   - encryptedPayload: OpenSSL-compatible encrypted data
//   - passphrase: decryption password/key
//
// Returns:
//   - decrypted plaintext
//   - error if decryption fails (wrong key, corrupted data, etc)
func DecryptAES(encryptedPayload, passphrase []byte) ([]byte, error) {
	// Validate minimum payload size
	const minPayloadSize = 16 // header(8) + salt(8)
	if len(encryptedPayload) < minPayloadSize {
		return nil, fmt.Errorf("encrypted payload too short: got %d bytes, need at least %d",
			len(encryptedPayload), minPayloadSize)
	}

	// Verify OpenSSL header
	opensslHeader := []byte(OpenSSLSaltedHeader)
	header := encryptedPayload[:8]
	if !bytes.Equal(header, opensslHeader) {
		return nil, fmt.Errorf("invalid format: OpenSSL header 'Salted__' not found")
	}

	// Extract salt and ciphertext with nonce
	salt := encryptedPayload[8:16]
	ciphertextWithNonce := encryptedPayload[16:]

	// Derive key from passphrase and salt
	key := deriveKey(passphrase, salt)

	// Initialize AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM for authenticated decryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce from beginning of ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertextWithNonce) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short: got %d bytes, need at least %d for nonce",
			len(ciphertextWithNonce), nonceSize)
	}

	nonce := ciphertextWithNonce[:nonceSize]
	ciphertext := ciphertextWithNonce[nonceSize:]

	// Decrypt and verify authentication tag
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong key or corrupted data): %w", err)
	}

	return plaintext, nil
}
