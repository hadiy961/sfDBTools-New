// File : internal/app/profile/helpers/key_resolver.go
// Deskripsi : Key resolution helper (extracted for P2 to avoid import cycles)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package helpers

import (
	"strings"

	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/shared/consts"
)

// ResolveProfileEncryptionKey resolves encryption key dari flag/env/prompt
func ResolveProfileEncryptionKey(existing string, allowPrompt bool) (key string, source string, err error) {
	if k := strings.TrimSpace(existing); k != "" {
		return k, "flag/state", nil
	}

	// Prefer TARGET key env, fallback ke SOURCE untuk kompatibilitas.
	if v, err := crypto.ResolveEnvSecret(consts.ENV_TARGET_PROFILE_KEY); err != nil {
		return "", "env", err
	} else if strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v), "env", nil
	}
	if v, err := crypto.ResolveEnvSecret(consts.ENV_SOURCE_PROFILE_KEY); err != nil {
		return "", "env", err
	} else if strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v), "env", nil
	}

	if !allowPrompt {
		return "", "", shared.NonInteractiveProfileKeyRequiredError()
	}

	// Prompt (interactive). crypto.ResolveKey akan mencoba env var yang diberikan dulu.
	k, src, e := crypto.ResolveKey("", consts.ENV_TARGET_PROFILE_KEY, true)
	if e != nil {
		return "", src, e
	}
	return strings.TrimSpace(k), src, nil
}
