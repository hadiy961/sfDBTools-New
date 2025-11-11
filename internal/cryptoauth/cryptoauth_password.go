package cryptoauth

// File : internal/cryptoauth/cryptoauth_password.go
// Deskripsi : Password authentication untuk crypto commands
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-11
// Last Modified : 2025-11-11

import (
	"fmt"
	"os"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/input"
)

// ValidatePassword meminta password dari user dan memvalidasi dengan ENV_PASSWORD_APP.
// Jika password salah, akan retry hingga benar atau user cancel (Ctrl+C).
//
// Return:
//   - error: error jika user cancel atau terjadi kesalahan sistem
func ValidatePassword() error {
	expectedPassword := consts.ENV_PASSWORD_APP

	for {
		// Gunakan input.AskPassword dari pkg/input
		password, err := input.AskPassword("Masukkan password untuk crypto utilities", nil)
		if err != nil {
			return fmt.Errorf("gagal membaca password: %w", err)
		}

		if password == expectedPassword {
			return nil
		}

		fmt.Println("✗ Password salah! Coba lagi atau tekan Ctrl+C untuk batal.")
		fmt.Println()
	}
}

// MustValidatePassword adalah wrapper untuk ValidatePassword yang akan exit jika validasi gagal.
// Digunakan untuk command yang membutuhkan password mandatory.
func MustValidatePassword() {
	if err := ValidatePassword(); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Autentikasi gagal: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Autentikasi berhasil")
	fmt.Println()
}
