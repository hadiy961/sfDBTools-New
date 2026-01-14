// File : internal/app/profile/helpers/key_resolver.go
// Deskripsi : Key resolution helper (extracted for P2 to avoid import cycles)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package helpers

import (
	"sfdbtools/internal/app/profile/helpers/keys"
)

// ResolveProfileEncryptionKey resolves encryption key dari flag/env/prompt
func ResolveProfileEncryptionKey(existing string, allowPrompt bool) (key string, source string, err error) {
	return keys.ResolveProfileEncryptionKey(existing, allowPrompt)
}
