// File : pkg/encrypt/env_values.go
// Deskripsi : Helper untuk encode/decode nilai ENV terenkripsi
// Author : Hadiyatna Muflihun
// Tanggal : 6 Januari 2026
// Last Modified : 6 Januari 2026

package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"sfdbtools/pkg/consts"
)

var (
	// Disimpan dalam bentuk obfuscated agar tidak muncul sebagai plaintext di binary (mis. saat menjalankan `strings`).
	envEncryptedPrefix = deobfuscateXORString([]byte{0xF9, 0xEC, 0xEE, 0xE8, 0xFE, 0xE5, 0xE5, 0xE6, 0xF9, 0x90}, 0xAA)
	envValueAADBytes   = deobfuscateXORBytes([]byte{0xD9, 0xCC, 0xCE, 0xC8, 0xDE, 0xC5, 0xC5, 0xC6, 0xD9, 0x87, 0xCF, 0xC4, 0xDC, 0x87, 0xDC, 0x9B}, 0xAA)
)

const (
	// envValueVersion adalah versi format payload.
	envValueVersion byte = 1

	// defaultMariaDBKeyFile adalah lokasi default file key material MariaDB.
	defaultMariaDBKeyFile = "/var/lib/mysql/key_maria_nbc.txt"
)

// EnvEncryptedPrefixForDisplay mengembalikan prefix env terenkripsi untuk ditampilkan di UI/help.
// Prefix ini sengaja tidak ditulis sebagai string literal supaya tidak muncul sebagai plaintext di binary.
func EnvEncryptedPrefixForDisplay() string {
	return envEncryptedPrefix
}

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

// EncodeEnvValue mengenkripsi plaintext menjadi format "prefix:<payload>".
// Payload menggunakan base64.RawURLEncoding (tanpa padding '=') dan format biner:
// [1 byte version][12 byte nonce][ciphertext+tag].
func EncodeEnvValue(plaintext string) (string, error) {
	plain := strings.TrimSpace(plaintext)
	if plain == "" {
		return "", errors.New("plaintext kosong")
	}

	key := deriveEnvMasterKey(defaultMariaDBKeyFile)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plain), envValueAADBytes)

	buf := make([]byte, 0, 1+len(nonce)+len(ciphertext))
	buf = append(buf, envValueVersion)
	buf = append(buf, nonce...)
	buf = append(buf, ciphertext...)

	payload := base64.RawURLEncoding.EncodeToString(buf)
	return envEncryptedPrefix + payload, nil
}

// DecodeEnvValue mendekode string. Jika tidak memakai prefix, nilai dikembalikan apa adanya.
// Jika memakai prefix dan payload invalid, akan mengembalikan error (fail-fast).
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
		return "", true, err
	}
	if len(raw) < 1 {
		return "", true, errors.New("payload terlalu pendek")
	}
	if raw[0] != envValueVersion {
		return "", true, errors.New("versi payload tidak dikenali")
	}

	key := deriveEnvMasterKey(defaultMariaDBKeyFile)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", true, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", true, err
	}

	need := 1 + gcm.NonceSize() + gcm.Overhead()
	if len(raw) < need {
		return "", true, errors.New("payload terlalu pendek")
	}

	nonceStart := 1
	nonceEnd := nonceStart + gcm.NonceSize()
	nonce := raw[nonceStart:nonceEnd]
	ciphertext := raw[nonceEnd:]

	plain, err := gcm.Open(nil, nonce, ciphertext, envValueAADBytes)
	if err != nil {
		return "", true, errors.New("gagal decrypt payload (kemungkinan master key berbeda/berubah; cek akses ke /var/lib/mysql/key_maria_nbc.txt dan pastikan proses encode/decode memakai kondisi yang konsisten): " + err.Error())
	}
	return string(plain), true, nil
}

// ResolveEnvSecret mengambil nilai env var, dan jika nilainya memakai prefix "prefix:" maka akan auto-decrypt. Jika payload invalid, mengembalikan error.
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
		return "", errors.New("env " + envVar + ": payload env terenkripsi tidak valid: " + err.Error())
	}
	if wasEncrypted {
		return decoded, nil
	}
	return raw, nil
}

func deriveEnvMasterKey(mariaDBKeyFilePath string) []byte {
	material := []byte(consts.ENV_PASSWORD_APP)
	if extra := readMariaDBKeyMaterial(mariaDBKeyFilePath); len(extra) > 0 {
		material = append(material, 0)
		material = append(material, extra...)
	}
	sum := sha256.Sum256(material)
	return sum[:]
}

func readMariaDBKeyMaterial(filePath string) []byte {
	if strings.TrimSpace(filePath) == "" {
		return nil
	}

	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

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
		if id == 1 || id == 100 {
			continue
		}
		if hexStr == "" {
			continue
		}

		// Normalisasi: buang spasi dan prefix 0x jika ada
		hexStr = strings.TrimPrefix(strings.ToLower(hexStr), "0x")
		hexStr = strings.ReplaceAll(hexStr, " ", "")
		keyBytes, err := hex.DecodeString(hexStr)
		if err != nil || len(keyBytes) == 0 {
			continue
		}
		candidates = append(candidates, candidate{id: id, key: keyBytes})
	}

	if len(candidates) == 0 {
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool { return candidates[i].id < candidates[j].id })
	return candidates[len(candidates)-1].key
}
