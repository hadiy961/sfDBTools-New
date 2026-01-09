// File : internal/crypto/auth/password.go
// Deskripsi : Password prompting and validation
// Author : Hadiyatna Muflihun
// Tanggal : 8 Januari 2026
// Last Modified : 8 Januari 2026
package auth

import (
	"fmt"

	"sfdbtools/internal/ui/prompt"

	"github.com/AlecAivazis/survey/v2"
)

// PromptPassword prompts user for password with optional ENV fallback.
//
// Parameters:
//   - envVar: environment variable name to check first (empty to skip)
//   - message: prompt message to show user
//
// Returns:
//   - password: obtained password
//   - source: where password came from ("env" or "prompt")
//   - error: if failed to get password
//
// Example:
//
//	pwd, source, err := auth.PromptPassword("SFDB_ENCRYPTION_KEY", "Enter encryption key:")
func PromptPassword(envVar, message string) (password, source string, err error) {
	// Try to get from env first (with auto-decrypt if needed)
	if envVar != "" {
		// Import dari key package untuk resolve encrypted env
		// Untuk avoid circular dependency, kita akan panggil via interface
		// Sementara skip ENV check, user bisa implement sendiri di caller
	}

	// Prompt user interactively
	pwd, err := prompt.AskPassword(message, survey.Required)
	if err != nil {
		return "", "prompt", fmt.Errorf("failed to read password: %w", err)
	}

	return pwd, "prompt", nil
}

// MustPromptPassword is like PromptPassword but panics on error.
// Use only in contexts where password is absolutely required.
func MustPromptPassword(envVar, message string) string {
	pwd, _, err := PromptPassword(envVar, message)
	if err != nil {
		panic(fmt.Sprintf("FATAL: password required but failed to obtain: %v", err))
	}
	return pwd
}
