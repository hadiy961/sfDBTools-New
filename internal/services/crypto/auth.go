package crypto

import (
	"fmt"
	"os"

	"sfDBTools/internal/ui/print"
	"sfDBTools/internal/ui/prompt"
	"sfDBTools/pkg/consts"
)

// ValidatePassword meminta password dari user dan memvalidasi dengan ENV_PASSWORD_APP.
// Jika password salah, akan retry hingga benar atau user cancel (Ctrl+C).
func ValidatePassword() error {
	expectedPassword := consts.ENV_PASSWORD_APP

	for {
		password, err := prompt.AskPassword("Masukkan password untuk crypto utilities", nil)
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

// MustValidatePassword adalah wrapper untuk ValidatePassword yang akan exit jika validasi gagal.
// Digunakan untuk command yang membutuhkan password mandatory.
func MustValidatePassword() {
	if err := ValidatePassword(); err != nil {
		print.PrintError(fmt.Sprintf("Autentikasi gagal: %v", err))
		os.Exit(1)
	}
	print.PrintSuccess("Autentikasi berhasil")
	print.Println()
}
