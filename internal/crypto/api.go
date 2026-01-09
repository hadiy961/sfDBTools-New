// Package crypto provides unified cryptographic operations for sfdbtools.
//
// This package consolidates all encryption, decryption, key management,
// and authentication functions into a single, well-organized API.
//
// # Architecture
//
//   - core/: AES-256-GCM primitives
//   - stream/: Streaming encryption/decryption for large files
//   - key/: Key resolution, derivation, and ENV secret handling
//   - auth/: Password prompting and validation
//   - file/: File encryption convenience functions
//
// # Data Encryption Example
//
//	ciphertext, err := crypto.EncryptData(plaintext, []byte(password))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Streaming Encryption Example
//
//	encryptor, err := crypto.NewStreamEncryptor(fileWriter, []byte(password))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer encryptor.Close()
//	io.Copy(encryptor, dataSource)
//
// # Key Resolution Example
//
//	key, source, err := crypto.ResolveKey(flagKey, "SFDB_ENCRYPTION_KEY", true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	log.Printf("Using key from: %s", source)
//
// # File Encryption Example
//
//	err := crypto.EncryptFile("data.txt", "data.txt.enc", []byte(password))
//
// # ENV Secret Example
//
//	// Encode secret to encrypted ENV value
//	encoded, _ := crypto.EncodeEnvSecret("my-secret-password")
//	// Output: SFENC:<base64-payload>
//
//	// Later, decode automatically
//	secret, _ := crypto.ResolveEnvSecret("MY_SECRET_VAR")
//
// File : internal/crypto/api.go
// Deskripsi : Public API facade untuk semua crypto operations
// Author : Hadiyatna Muflihun
// Tanggal : 8 Januari 2026
// Last Modified : 8 Januari 2026
package crypto

import (
	"io"

	"sfdbtools/internal/crypto/auth"
	"sfdbtools/internal/crypto/core"
	"sfdbtools/internal/crypto/file"
	"sfdbtools/internal/crypto/key"
	"sfdbtools/internal/crypto/stream"
)

// ========================
// Data Encryption/Decryption
// ========================

// EncryptData encrypts data using AES-256-GCM (OpenSSL compatible).
//
// Returns encrypted payload with "Salted__" header, salt, nonce, and ciphertext.
// Compatible with: openssl enc -aes-256-gcm -pbkdf2 -iter 600000
func EncryptData(data, passphrase []byte) ([]byte, error) {
	return core.EncryptAES(data, passphrase)
}

// DecryptData decrypts AES-256-GCM encrypted data (OpenSSL compatible).
//
// Expects encrypted payload with "Salted__" header from EncryptData.
// Returns error if wrong key or corrupted data.
func DecryptData(ciphertext, passphrase []byte) ([]byte, error) {
	return core.DecryptAES(ciphertext, passphrase)
}

// ========================
// Streaming Encryption/Decryption
// ========================

// NewStreamEncryptor creates a streaming encryption writer.
//
// Data written to returned writer is encrypted in 64KB chunks and written
// to underlying writer. Must call Close() to finalize encryption.
//
// Returns io.WriteCloser for writing plaintext data.
func NewStreamEncryptor(w io.Writer, passphrase []byte) (io.WriteCloser, error) {
	return stream.NewWriter(w, passphrase)
}

// NewStreamDecryptor creates a streaming decryption reader.
//
// Reads encrypted data from source reader and returns plaintext.
// Compatible with output from NewStreamEncryptor.
//
// Returns io.Reader for reading decrypted data.
func NewStreamDecryptor(r io.Reader, passphrase string) (io.Reader, error) {
	return stream.NewReader(r, passphrase)
}

// ========================
// File Encryption/Decryption
// ========================

// EncryptFile encrypts a file using streaming encryption.
//
// Reads from inputPath, encrypts, and writes to outputPath.
// Uses streaming to handle large files efficiently.
func EncryptFile(inputPath, outputPath string, passphrase []byte) error {
	return file.Encrypt(inputPath, outputPath, passphrase)
}

// DecryptFile decrypts a file using streaming decryption.
//
// Reads encrypted file from inputPath, decrypts, and writes to outputPath.
// Returns error if wrong key or corrupted data.
func DecryptFile(inputPath, outputPath string, passphrase []byte) error {
	return file.Decrypt(inputPath, outputPath, passphrase)
}

// ========================
// Key Management
// ========================

// ResolveKey resolves encryption key from multiple sources (priority order):
//  1. Explicit key from flag/parameter
//  2. Environment variable (with auto-decrypt if encrypted)
//  3. Interactive prompt (if allowPrompt = true)
//
// Returns:
//   - key: resolved encryption key
//   - source: where key came from ("flag", "env", "prompt")
//   - error: if key cannot be obtained
func ResolveKey(explicit, envName string, allowPrompt bool) (keyStr, source string, err error) {
	return key.Resolve(explicit, envName, allowPrompt)
}

// ResolveKeyOrFail is like ResolveKey but panics if key cannot be obtained.
// Use only in contexts where key is absolutely required.
func ResolveKeyOrFail(explicit, envName string) string {
	return key.ResolveOrFail(explicit, envName)
}

// ========================
// ENV Secret Management
// ========================

// EncodeEnvSecret encodes plaintext to encrypted ENV value format.
//
// Returns string with prefix: "SFENC:<base64-payload>"
// Uses master key derived from MariaDB key file.
//
// Example output: SFENC:AQxBMjM0NTY3ODkwMTI...
func EncodeEnvSecret(plaintext string) (string, error) {
	return key.EncodeEnvValue(plaintext)
}

// DecodeEnvSecret decodes encrypted ENV value.
//
// If value doesn't have SFENC prefix, returns as-is.
// If value has prefix but invalid payload, returns error.
//
// Returns:
//   - decoded: decrypted value or original value
//   - wasEncrypted: whether value was encrypted
//   - error: if decryption fails
func DecodeEnvSecret(value string) (decoded string, wasEncrypted bool, err error) {
	return key.DecodeEnvValue(value)
}

// ResolveEnvSecret gets ENV var value and auto-decrypts if encrypted.
//
// Returns empty string if ENV var not set.
// Returns error if value is encrypted but decryption fails.
func ResolveEnvSecret(envName string) (string, error) {
	return key.ResolveEnvSecret(envName)
}

// EncryptedPrefixForDisplay returns the prefix for encrypted ENV values.
// Used in CLI help text and documentation.
func EncryptedPrefixForDisplay() string {
	return key.EncryptedPrefixForDisplay()
}

// ========================
// MariaDB Key Management
// ========================

// SetMariaDBKeyFilePath sets the path to MariaDB key file.
// Used for deriving master key for ENV secret encryption.
//
// Default: /var/lib/mysql/key_maria_nbc.txt
//
// Should be called once at startup from config.yaml.
func SetMariaDBKeyFilePath(path string) {
	key.SetMariaDBKeyFilePath(path)
}

// GetMariaDBKeyFilePath returns the effective MariaDB key file path.
func GetMariaDBKeyFilePath() string {
	return key.GetMariaDBKeyFilePath()
}

// ========================
// Authentication
// ========================

// PromptPassword prompts user for password with optional ENV fallback.
//
// Parameters:
//   - envVar: environment variable to check first (empty to skip)
//   - message: prompt message to show user
//
// Returns:
//   - password: obtained password
//   - source: where password came from ("env" or "prompt")
//   - error: if failed to get password
func PromptPassword(envVar, message string) (password, source string, err error) {
	return auth.PromptPassword(envVar, message)
}

// ValidateApplicationPassword validates user with application password.
//
// Prompts for password and validates against configured app password.
// Retries on wrong password until correct or user cancels.
func ValidateApplicationPassword() error {
	return auth.ValidateApplicationPassword()
}

// MustValidateApplicationPassword validates or exits with code 1.
//
// Use for commands that require mandatory authentication.
func MustValidateApplicationPassword() {
	auth.MustValidateApplicationPassword()
}

// ========================
// Monitoring/Stats
// ========================

// GetFailedDecodeCount returns the number of failed ENV decode attempts.
// Used for monitoring/alerting suspicious activity.
func GetFailedDecodeCount() uint64 {
	return key.GetFailedDecodeCount()
}
