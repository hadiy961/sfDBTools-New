// File : internal/profile/profile_key.go
// Deskripsi : Helper untuk resolve kunci enkripsi profil (flag/env/prompt)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 6 Januari 2026

package profile

import (
	"fmt"
	"strings"

	cryptokey "sfdbtools/internal/services/crypto/helpers"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/encrypt"
	"sfdbtools/internal/shared/validation"
)

func resolveProfileEncryptionKey(existing string, allowPrompt bool) (key string, source string, err error) {
	if k := strings.TrimSpace(existing); k != "" {
		return k, "flag/state", nil
	}

	// Prefer TARGET key env, fallback ke SOURCE untuk kompatibilitas.
	if v, err := encrypt.ResolveEnvSecret(consts.ENV_TARGET_PROFILE_KEY); err != nil {
		return "", "env", err
	} else if strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v), "env", nil
	}
	if v, err := encrypt.ResolveEnvSecret(consts.ENV_SOURCE_PROFILE_KEY); err != nil {
		return "", "env", err
	} else if strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v), "env", nil
	}

	if !allowPrompt {
		return "", "", fmt.Errorf(
			consts.ProfileErrNonInteractiveProfileKeyRequiredFmt,
			consts.ENV_TARGET_PROFILE_KEY,
			consts.ENV_SOURCE_PROFILE_KEY,
			validation.ErrNonInteractive,
		)
	}

	// Prompt (interactive). cryptokey.ResolveEncryptionKey akan mencoba env var yang diberikan dulu.
	k, src, e := cryptokey.ResolveEncryptionKey("", consts.ENV_TARGET_PROFILE_KEY)
	if e != nil {
		return "", src, e
	}
	return strings.TrimSpace(k), src, nil
}
