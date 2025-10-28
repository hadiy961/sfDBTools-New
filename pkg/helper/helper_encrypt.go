package helper

import (
	"fmt"
	"sfDBTools/pkg/encrypt"
	"strings"
)

// ResolveEncryptionKey mengembalikan kunci enkripsi final dan sumbernya.
func ResolveEncryptionKey(existing string, env string) (string, string, error) {
	if k := strings.TrimSpace(existing); k != "" {
		return k, "flag/state", nil
	}
	// Jika tidak ada existing, minta dari env atau prompt
	pwd, source, err := encrypt.EncryptionPrompt("Masukkan Kunci Enkripsi :", env)
	if err != nil {
		return "", source, err
	}
	// Validasi tambahan: pastikan tidak kosong setelah trim
	pwd = strings.TrimSpace(pwd)
	if pwd == "" {
		return "", source, fmt.Errorf("kunci enkripsi kosong")
	}
	return pwd, source, nil
}
