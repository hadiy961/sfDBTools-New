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
// If password is wrong, prompts again until correct or user cancels (Ctrl+C).
// Used by crypto utilities to ensure authorized access.
//
// Returns error if user cancels or read fails.
func ValidateApplicationPassword() error {
	expectedPassword := consts.ENV_PASSWORD_APP

	for {
		password, _, err := PromptPassword("", "Masukkan password untuk crypto utilities")
		if err != nil {
			return fmt.Errorf("gagal membaca password: %w", err)
		}

		if password == expectedPassword {
			return nil
		}

		print.PrintError("Password salah! Coba lagi atau tekan Ctrl+C untuk batal.")
		print.Println()
	}
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
