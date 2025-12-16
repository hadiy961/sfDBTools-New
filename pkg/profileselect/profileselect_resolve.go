package profileselect

import "sfDBTools/pkg/encrypt"

// GetEncryptionPassword mendapatkan password enkripsi dari env atau prompt.
func GetEncryptionPassword(env string) (string, string, error) {
	return encrypt.PromptPassword(env, "🔑 Encryption password: ")
}
