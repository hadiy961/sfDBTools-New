// File : pkg/validation/validation.go
// Deskripsi : Fungsi fungsi untuk validasi ada disini
// Author : Hadiyatna Muflihun
// Tanggal : 11 November 2025
// Last Modified : 11 November 2025
package validation

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sfDBTools/pkg/consts"
	"strings"

	"github.com/AlecAivazis/survey/v2/terminal"
)

// ValidateSubdirPattern memastikan pola valid:
// - jumlah '{' sama dengan jumlah '}'
// - token hanya mengandung nama yang diizinkan (builtin atau vars)
// - tidak memulai dengan '/' (absolute path)
// - tidak mengandung path traversal '..'
func ValidateSubdirPattern(pattern string, vars map[string]string) error {
	if strings.Count(pattern, "{") != strings.Count(pattern, "}") {
		return fmt.Errorf("kurung kurawal tidak seimbang pada pola")
	}
	if pattern == "" {
		return fmt.Errorf("pola kosong")
	}
	if strings.HasPrefix(pattern, string(os.PathSeparator)) {
		return fmt.Errorf("pola tidak boleh absolut (tidak boleh diawali dengan %q)", string(os.PathSeparator))
	}
	if strings.Contains(pattern, "..") {
		return fmt.Errorf("pola tidak boleh mengandung path traversal '..'")
	}

	// allowed builtin tokens
	builtins := map[string]bool{
		"date":      true,
		"timestamp": true,
		"year":      true,
		"month":     true,
		"day":       true,
	}

	// regex menangkap {name} atau {name:format}
	re := regexp.MustCompile(`\{([a-zA-Z0-9_]+)(?::([^}]+))?\}`)
	matches := re.FindAllStringSubmatch(pattern, -1)
	for _, m := range matches {
		if len(m) < 2 {
			return fmt.Errorf("format token tidak valid")
		}
		name := m[1]
		if builtins[name] {
			continue
		}
		// jika bukan builtin, pastikan ada di vars
		if _, ok := vars[name]; !ok {
			return fmt.Errorf("token tidak dikenal '%s' di pola; token yang diizinkan: date,timestamp,year,month,day atau kunci yang tersedia di vars", name)
		}
	}
	return nil
}

var ErrUserCancelled = errors.New("user_cancelled")
var ErrConnectionFailedRetry = errors.New("connection_failed_retry")

// HandleInputError menangani error dari input/survey dan mengubahnya menjadi ErrUserCancelled jika perlu.
func HandleInputError(err error) error {
	if err == terminal.InterruptErr {
		return ErrUserCancelled
	}
	return err
}

// ProfileExt memastikan nama memiliki suffix .cnf.enc
func ProfileExt(name string) string {
	if strings.HasSuffix(name, consts.ExtCnfEnc) {
		return name
	}
	if strings.HasSuffix(name, consts.ExtCnf) {
		return name + consts.ExtEnc
	}
	return name + consts.ExtCnfEnc
}
