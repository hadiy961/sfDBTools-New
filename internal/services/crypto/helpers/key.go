package helpers

import (
	"fmt"
	"strings"

	"sfdbtools/internal/shared/encrypt"
)

// ResolveEncryptionKey mengembalikan kunci enkripsi final dan sumbernya.
func ResolveEncryptionKey(existing string, env string) (string, string, error) {
	if k := strings.TrimSpace(existing); k != "" {
		return k, "flag/state", nil
	}
	pwd, source, err := encrypt.PromptPassword(env, "Masukkan Kunci Enkripsi :")
	if err != nil {
		return "", source, err
	}
	if pwd = strings.TrimSpace(pwd); pwd == "" {
		return "", source, fmt.Errorf("kunci enkripsi kosong")
	}
	return pwd, source, nil
}
