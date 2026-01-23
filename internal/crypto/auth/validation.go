// File : internal/crypto/auth/validation.go
// Deskripsi : Application password validation
// Author : Hadiyatna Muflihun
// Tanggal : 8 Januari 2026
// Last Modified : 8 Januari 2026
package auth

import (
	"fmt"
	"os"

	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"
)

// ValidateApplicationPassword prompts user for application password and validates.
//
// Allows maximum 3 attempts. If password is wrong 3 times, returns error.
// Used by crypto utilities to ensure authorized access.
//
// Returns error if user cancels, read fails, or exceeds max attempts.
func ValidateApplicationPassword() error {
	expectedPassword := consts.ENV_PASSWORD_APP
	const maxAttempts = 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		password, _, err := PromptPassword("", "Masukkan password untuk crypto utilities")
		if err != nil {
			return fmt.Errorf("gagal membaca password: %w", err)
		}

		if password == expectedPassword {
			return nil
		}

		if attempt < maxAttempts {
			print.PrintError(fmt.Sprintf("Password salah! Percobaan %d/%d. Coba lagi atau tekan Ctrl+C untuk batal.", attempt, maxAttempts))
			print.Println()
		}
	}

	return fmt.Errorf("maksimum percobaan password tercapai (%d kali)", maxAttempts)
}

// MustValidateApplicationPassword validates application password or exits.
//
// Use for commands that require mandatory authentication.
// Exits with code 1 if validation fails.
func MustValidateApplicationPassword() {
	if err := ValidateApplicationPassword(); err != nil {
		print.PrintError(fmt.Sprintf("Autentikasi gagal: %v", err))
		os.Exit(1)
	}
	print.PrintSuccess("Autentikasi berhasil")
	print.Println()
}
