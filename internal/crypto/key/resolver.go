// File : internal/crypto/key/resolver.go
// Deskripsi : Key resolution dari multiple sources (flag, env, prompt)
// Author : Hadiyatna Muflihun
// Tanggal : 8 Januari 2026
// Last Modified : 9 Januari 2026
package key

import (
	"fmt"
	"strings"

	"sfdbtools/internal/ui/prompt"

	"github.com/AlecAivazis/survey/v2"
)

// Resolve resolves encryption key from multiple sources in priority order:
//  1. Explicit key (from flag/parameter)
//  2. Environment variable
//  3. Interactive prompt (if allowPrompt = true)
//
// Parameters:
//   - explicit: key from command-line flag or parameter
//   - envName: environment variable name to check
//   - allowPrompt: whether to prompt user if key not found
//
// Returns:
//   - key: resolved encryption key
//   - source: where key came from ("flag", "env", "prompt")
//   - error: if key cannot be obtained
//
// Example:
//
//	key, source, err := key.Resolve(opts.Key, "SFDB_ENCRYPTION_KEY", true)
func Resolve(explicit, envName string, allowPrompt bool) (key, source string, err error) {
	// Priority 1: Explicit key from flag/parameter
	if k := strings.TrimSpace(explicit); k != "" {
		return k, "flag", nil
	}

	// Priority 2: Environment variable (with auto-decrypt if encrypted)
	if envName != "" {
		if envVal, err := ResolveEnvSecret(envName); err != nil {
			return "", "env", fmt.Errorf("failed to resolve env %s: %w", envName, err)
		} else if envVal != "" {
			return envVal, "env", nil
		}
	}

	// Priority 3: Interactive prompt (if allowed)
	if allowPrompt {
		return promptForKey(envName)
	}

	return "", "", fmt.Errorf("encryption key not found: provide via flag or environment variable %s", envName)
}

// ResolveOrFail is like Resolve but panics if key cannot be obtained.
// Use only in contexts where key is absolutely required.
func ResolveOrFail(explicit, envName string) string {
	key, _, err := Resolve(explicit, envName, true)
	if err != nil {
		panic(fmt.Sprintf("FATAL: encryption key required but not found: %v", err))
	}
	return key
}

// promptForKey prompts user for encryption key interactively.
func promptForKey(envName string) (key, source string, err error) {
	msg := "Masukkan kunci enkripsi: "
	if strings.TrimSpace(envName) != "" {
		msg = fmt.Sprintf("Masukkan kunci enkripsi (atau set ENV %s): ", strings.TrimSpace(envName))
	}

	k, err := prompt.AskPassword(msg, survey.Required)
	if err != nil {
		return "", "prompt", err
	}
	k = strings.TrimSpace(k)
	if k == "" {
		return "", "prompt", fmt.Errorf("kunci enkripsi kosong")
	}
	return k, "prompt", nil
}
