// File : internal/ui/input/input_validator.go
// Deskripsi : Fungsi utilitas untuk validasi input user
// Author : Hadiyatna Muflihun
// Tanggal : 3 Oktober 2024
// Last Modified : 5 Januari 2026
package input

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
)

// ValidatePort adalah validator untuk memastikan input adalah port yang valid.
func ValidatePort(ans interface{}) error {
	// survey v2 mengembalikan string untuk input, jadi kita perlu konversi
	str, ok := ans.(string)
	if !ok {
		return fmt.Errorf("tidak dapat memvalidasi tipe data non-string")
	}

	port, err := strconv.Atoi(str)
	if err != nil {
		return fmt.Errorf("port harus berupa angka")
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("nomor port harus di antara 1 dan 65535")
	}
	return nil
}

// ValidateFilename adalah validator untuk memastikan input aman sebagai nama file.
func ValidateFilename(ans interface{}) error {
	str, ok := ans.(string)
	if !ok {
		return fmt.Errorf("tidak dapat memvalidasi tipe data non-string")
	}

	// Hanya izinkan huruf, angka, underscore, dan hyphen
	// ^[a-zA-Z]      -> Harus diawali dengan satu huruf (besar atau kecil).
	// [a-zA-Z0-9_-]* -> Diikuti oleh 0 atau lebih huruf, angka, underscore, atau hyphen.
	// $              -> Sampai akhir string.
	re := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !re.MatchString(str) {
		return fmt.Errorf("nama harus diawali dengan huruf dan hanya boleh berisi huruf, angka, underscore (_), dan hyphen (-)")
	}
	return nil
}

// ComposeValidators menggabungkan beberapa validator menjadi satu.
// Jika salah satu gagal, ia akan mengembalikan error pertama.
func ComposeValidators(validators ...survey.Validator) survey.Validator {
	return func(val interface{}) error {
		for _, v := range validators {
			if err := v(val); err != nil {
				return err
			}
		}
		return nil
	}
}
