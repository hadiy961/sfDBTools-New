// File : internal/app/profile/wizard/validators.go
// Deskripsi : Validator reusable untuk prompt profile (DB & SSH)
// Author : Hadiyatna Muflihun
// Tanggal : 9 Januari 2026
// Last Modified : 9 Januari 2026

package wizard

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func validateNoLeadingTrailingSpaces(fieldName string) survey.Validator {
	return func(ans interface{}) error {
		s, _ := ans.(string)
		if strings.TrimSpace(s) != s {
			if fieldName == "" {
				return fmt.Errorf("input tidak boleh diawali/diakhiri spasi")
			}
			return fmt.Errorf("%s tidak boleh diawali/diakhiri spasi", fieldName)
		}
		return nil
	}
}

func validateOptionalNoLeadingTrailingSpaces(fieldName string) survey.Validator {
	return func(ans interface{}) error {
		s, _ := ans.(string)
		// kosong (atau hanya spasi) dianggap "tidak mengisi" pada field opsional.
		if strings.TrimSpace(s) == "" {
			return nil
		}
		if strings.TrimSpace(s) != s {
			if fieldName == "" {
				return fmt.Errorf("input tidak boleh diawali/diakhiri spasi")
			}
			return fmt.Errorf("%s tidak boleh diawali/diakhiri spasi", fieldName)
		}
		return nil
	}
}

func validateNoControlChars(fieldName string) survey.Validator {
	return func(ans interface{}) error {
		s, _ := ans.(string)
		if strings.ContainsAny(s, "\n\r\t") {
			if fieldName == "" {
				return fmt.Errorf("input tidak boleh mengandung karakter kontrol")
			}
			return fmt.Errorf("%s tidak boleh mengandung karakter kontrol", fieldName)
		}
		return nil
	}
}

func validateOptionalNoControlChars(fieldName string) survey.Validator {
	return func(ans interface{}) error {
		s, _ := ans.(string)
		if strings.TrimSpace(s) == "" {
			return nil
		}
		return validateNoControlChars(fieldName)(ans)
	}
}

func validateNoSpaces(fieldName string) survey.Validator {
	return func(ans interface{}) error {
		s, _ := ans.(string)
		if strings.ContainsAny(s, " \t") {
			if fieldName == "" {
				return fmt.Errorf("input tidak boleh mengandung spasi")
			}
			return fmt.Errorf("%s tidak boleh mengandung spasi", fieldName)
		}
		return nil
	}
}

func validateNotBlank(fieldName string) survey.Validator {
	return func(ans interface{}) error {
		s, _ := ans.(string)
		if strings.TrimSpace(s) == "" {
			if fieldName == "" {
				return fmt.Errorf("input tidak boleh kosong")
			}
			return fmt.Errorf("%s tidak boleh kosong", fieldName)
		}
		return nil
	}
}

func validatePortRange(min int, max int, allowZero bool, fieldName string) survey.Validator {
	return func(ans interface{}) error {
		s, _ := ans.(string)
		trim := strings.TrimSpace(s)
		if trim == "" {
			return fmt.Errorf("%s tidak boleh kosong", fieldName)
		}
		v, err := strconv.Atoi(trim)
		if err != nil {
			return fmt.Errorf("%s harus berupa angka", fieldName)
		}
		if allowZero && v == 0 {
			return nil
		}
		if v < min || v > max {
			if allowZero {
				return fmt.Errorf("%s harus 0 (otomatis) atau %d-%d", fieldName, min, max)
			}
			return fmt.Errorf("%s harus %d-%d", fieldName, min, max)
		}
		return nil
	}
}

func validateOptionalExistingFilePath(fieldName string) survey.Validator {
	return func(ans interface{}) error {
		s, _ := ans.(string)
		p := strings.TrimSpace(s)
		if p == "" {
			return nil
		}
		if !filepath.IsAbs(p) {
			if wd, err := os.Getwd(); err == nil {
				p = filepath.Join(wd, p)
			}
		}
		p = filepath.Clean(p)
		fi, err := os.Stat(p)
		if err != nil {
			return fmt.Errorf("%s tidak bisa diakses: %s", fieldName, p)
		}
		if fi.IsDir() {
			return fmt.Errorf("%s tidak valid (path adalah direktori): %s", fieldName, p)
		}
		return nil
	}
}
