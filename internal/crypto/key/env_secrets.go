// File : internal/crypto/key/env_secrets.go
// Deskripsi : ENV value encryption/decryption dengan master key derivation
// Author : Hadiyatna Muflihun
// Tanggal : 8 Januari 2026
// Last Modified : 8 Januari 2026
package key

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"sfdbtools/internal/shared/consts"
)

const (
	// envValueVersion adalah versi format payload
	envValueVersion byte = 1

	// defaultMariaDBKeyFile adalah lokasi default file key material MariaDB
	defaultMariaDBKeyFile = "/var/lib/mysql/key_maria_nbc.txt"
)

var (
	// Obfuscated prefix untuk avoid plaintext di binary
	envEncryptedPrefix = deobfuscateXORString([]byte{0xF9, 0xEC, 0xEE, 0xE8, 0xFE, 0xE5, 0xE5, 0xE6, 0xF9, 0x90}, 0xAA)
	envValueAADBytes   = deobfuscateXORBytes([]byte{0xD9, 0xCC, 0xCE, 0xC8, 0xDE, 0xC5, 0xC5, 0xC6, 0xD9, 0x87, 0xCF, 0xC4, 0xDC, 0x87, 0xDC, 0x9B}, 0xAA)

	// Cache untuk MariaDB key material (thread-safe)
	mariaDBKeyCache     []byte
	mariaDBKeyCacheOnce sync.Once
	mariaDBKeyCachePath string

	// Counter untuk failed decode attempts (monitoring/alerting)
	failedDecodeCount uint64

	// Path file key material MariaDB (configurable)
	mariaDBKeyFilePath = defaultMariaDBKeyFile
)

// SetMariaDBKeyFilePath sets the path to MariaDB key file for master key derivation.
// If empty, uses default: /var/lib/mysql/key_maria_nbc.txt
//
// This should be called once at startup from config.yaml.
func SetMariaDBKeyFilePath(path string) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		mariaDBKeyFilePath = defaultMariaDBKeyFile
		return
	}
	mariaDBKeyFilePath = trimmed
}

// GetMariaDBKeyFilePath returns the effective MariaDB key file path.
func GetMariaDBKeyFilePath() string {
	if strings.TrimSpace(mariaDBKeyFilePath) == "" {
		return defaultMariaDBKeyFile
	}
	return mariaDBKeyFilePath
}

// EncryptedPrefixForDisplay returns the prefix for encrypted ENV values.
// Used in CLI help text and documentation.
func EncryptedPrefixForDisplay() string {
	return envEncryptedPrefix
}

// GetFailedDecodeCount returns the number of failed decode attempts.
// Used for monitoring/alerting suspicious activity.
func GetFailedDecodeCount() uint64 {
	return atomic.LoadUint64(&failedDecodeCount)
}

// EncodeEnvValue encrypts plaintext to format "prefix:<payload>".
//
// Payload format (base64.RawURLEncoding):
//
//	[1 byte version][12 byte nonce][ciphertext+tag]
//
// Uses master key derived from MariaDB key file.
func EncodeEnvValue(plaintext string) (string, error) {
	plain := strings.TrimSpace(plaintext)
	if plain == "" {
		return "", errors.New("plaintext kosong")
	}

	key := deriveEnvMasterKey(GetMariaDBKeyFilePath())
	defer zeroBytes(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plain), envValueAADBytes)

	// Build payload: version + nonce + ciphertext
	buf := make([]byte, 0, 1+len(nonce)+len(ciphertext))
	buf = append(buf, envValueVersion)
	buf = append(buf, nonce...)
	buf = append(buf, ciphertext...)

	payload := base64.RawURLEncoding.EncodeToString(buf)
	return envEncryptedPrefix + payload, nil
}

// DecodeEnvValue decodes encrypted ENV value.
//
// If value doesn't have encrypted prefix, returns as-is (wasEncrypted=false).
// If value has prefix but invalid payload, returns error.
//
// Returns:
//   - decoded: decrypted value or original value
//   - wasEncrypted: whether value was encrypted
//   - error: if decryption fails
func DecodeEnvValue(value string) (decoded string, wasEncrypted bool, err error) {
	v := strings.TrimSpace(value)
	if !strings.HasPrefix(v, envEncryptedPrefix) {
		return value, false, nil
	}

	payload := strings.TrimSpace(strings.TrimPrefix(v, envEncryptedPrefix))
	if payload == "" {
		return "", true, errors.New("payload kosong")
	}

	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", true, errors.New("format payload tidak valid")
	}
	if len(raw) < 1 {
		return "", true, errors.New("payload terlalu pendek")
	}
	if raw[0] != envValueVersion {
		return "", true, fmt.Errorf("versi payload tidak dikenali: got %d, expected %d", raw[0], envValueVersion)
	}

	key := deriveEnvMasterKey(GetMariaDBKeyFilePath())
	defer zeroBytes(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", true, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", true, fmt.Errorf("failed to create GCM: %w", err)
	}

	need := 1 + gcm.NonceSize() + gcm.Overhead()
	if len(raw) < need {
		return "", true, fmt.Errorf("payload terlalu pendek: got %d bytes, need at least %d", len(raw), need)
	}

	nonceStart := 1
	nonceEnd := nonceStart + gcm.NonceSize()
	nonce := raw[nonceStart:nonceEnd]
	ciphertext := raw[nonceEnd:]

	plain, err := gcm.Open(nil, nonce, ciphertext, envValueAADBytes)
	if err != nil {
		// Log failed attempts for monitoring
		count := atomic.AddUint64(&failedDecodeCount, 1)
		if count%10 == 1 { // Log every 10 failures to avoid spam
			log.Printf("WARNING: Failed to decrypt env value (attempt #%d). Possible causes: different master key, wrong MariaDB key file, or corrupted payload.\n", count)
		}
		// Don't expose crypto error details (oracle attack prevention)
		return "", true, errors.New("gagal decrypt payload (kemungkinan master key berbeda/berubah; cek akses ke MariaDB key file)")
	}

	decodedStr := string(plain)
	defer zeroBytes(plain)
	return decodedStr, true, nil
}

// ResolveEnvSecret gets ENV var value and auto-decrypts if encrypted.
//
// Returns empty string if ENV var not set.
// Returns error if value is encrypted but decryption fails.
func ResolveEnvSecret(envVar string) (string, error) {
	if strings.TrimSpace(envVar) == "" {
		return "", errors.New("nama env var kosong")
	}

	raw := os.Getenv(envVar)
	if raw == "" {
		return "", nil
	}

	decoded, wasEncrypted, err := DecodeEnvValue(raw)
	if err != nil {
		return "", fmt.Errorf("env %s: payload terenkripsi tidak valid", envVar)
	}
	if wasEncrypted {
		return decoded, nil
	}
	return raw, nil
}

// Helper functions

func deobfuscateXORBytes(obfuscated []byte, key byte) []byte {
	out := make([]byte, len(obfuscated))
	for i := 0; i < len(obfuscated); i++ {
		out[i] = obfuscated[i] ^ key
	}
	return out
}

func deobfuscateXORString(obfuscated []byte, key byte) string {
	return string(deobfuscateXORBytes(obfuscated, key))
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func deriveEnvMasterKey(mariaDBKeyFilePath string) []byte {
	material := []byte(consts.ENV_PASSWORD_APP)

	// Add MariaDB key material if available
	extra := getCachedMariaDBKeyMaterial(mariaDBKeyFilePath)
	if len(extra) > 0 {
		material = append(material, 0)
		material = append(material, extra...)
	}

	sum := sha256.Sum256(material)
	return sum[:]
}

func getCachedMariaDBKeyMaterial(filePath string) []byte {
	if filePath != mariaDBKeyCachePath {
		// Path changed, reset cache
		mariaDBKeyCacheOnce = sync.Once{}
		mariaDBKeyCachePath = filePath
	}

	mariaDBKeyCacheOnce.Do(func() {
		mariaDBKeyCache = readMariaDBKeyMaterial(filePath)
	})
	return mariaDBKeyCache
}

func readMariaDBKeyMaterial(filePath string) []byte {
	if strings.TrimSpace(filePath) == "" {
		return nil
	}

	if stat, statErr := os.Stat(filePath); statErr == nil && stat != nil {
		b, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("WARNING: MariaDB key file %s exists but cannot be read (%v). Falling back to hardcoded seed only (weak security).\n", filePath, err)
			return nil
		}
		return parseMariaDBKeyMaterial(b)
	}
	// File doesn't exist (not an error, silent fallback)
	return nil
}

func parseMariaDBKeyMaterial(b []byte) []byte {
	lines := strings.Split(string(b), "\n")
	type candidate struct {
		id  int
		key []byte
	}
	candidates := make([]candidate, 0, 4)

	for _, line := range lines {
		ln := strings.TrimSpace(line)
		if ln == "" {
			continue
		}
		parts := strings.SplitN(ln, ";", 2)
		if len(parts) != 2 {
			continue
		}

		idStr := strings.TrimSpace(parts[0])
		hexStr := strings.TrimSpace(parts[1])

		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		// Skip reserved IDs
		if id == 1 || id == 100 {
			continue
		}
		if hexStr == "" {
			continue
		}

		// Normalize: remove spaces and 0x prefix
		hexStr = strings.TrimPrefix(strings.ToLower(hexStr), "0x")
		hexStr = strings.ReplaceAll(hexStr, " ", "")

		keyBytes, err := hex.DecodeString(hexStr)
		// Validate minimum key length: 16 bytes (128 bit)
		if err != nil || len(keyBytes) < 16 {
			continue
		}

		candidates = append(candidates, candidate{id: id, key: keyBytes})
	}

	if len(candidates) == 0 {
		return nil
	}

	// Use key with highest ID
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].id < candidates[j].id })
	return candidates[len(candidates)-1].key
}
