package profileselect

import (
	"fmt"
	"sfDBTools/pkg/encrypt"
)

// GetEncryptionPassword mendapatkan password enkripsi dari env atau prompt.
func GetEncryptionPassword(env string) (string, string, error) {
	// Dapatkan password enkripsi dari env atau prompt
	encryptionPassword, source, err := encrypt.EncryptionPrompt("🔑 Encryption password: ", env)
	if err != nil {
		return "", "", fmt.Errorf("failed to get encryption password: %w", err)
	}
	return encryptionPassword, source, nil
}
