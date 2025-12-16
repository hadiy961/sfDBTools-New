// File : pkg/input/input_general.go
// Deskripsi : Fungsi utilitas untuk input interaktif dari user
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package input

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

// Removed duplicate PrintError - use ui.PrintError instead

// --- Menu dan Interaktif Input ---

// ShowMenuAndClear displays a menu, gets user choice, then clears screen
// Menggantikan logika manual di terminal/helpers.go
func ShowMenu(title string, options []string) (int, error) {
	var selectedIndex int

	prompt := &survey.Select{
		Message: title,
		Options: options,
	}

	// survey.AskOne menangani tampilan menu dan navigasi.
	// Kita menggunakan terminal.Stdio untuk memastikan input/output standar.
	err := survey.AskOne(prompt, &selectedIndex, survey.WithStdio(
		os.Stdin,  // In
		os.Stdout, // Out
		os.Stderr, // Err
	))

	// Jika ada error (misalnya Ctrl+C)
	if err != nil {
		return 0, err
	}

	// Hasil Survey adalah 0-based index. Kembalikan 1-based.
	return selectedIndex + 1, nil
}

// ShowMultiSelect menampilkan menu multi-select dan mengembalikan indeks terpilih (1-based).
func ShowMultiSelect(title string, options []string) ([]int, error) {
	var selected []string

	prompt := &survey.MultiSelect{
		Message: title,
		Options: options,
	}

	err := survey.AskOne(prompt, &selected, survey.WithStdio(
		os.Stdin,
		os.Stdout,
		os.Stderr,
	))
	if err != nil {
		return nil, err
	}

	// map selected strings back to indices (1-based)
	idxs := make([]int, 0, len(selected))
	for _, sel := range selected {
		for i, opt := range options {
			if opt == sel {
				idxs = append(idxs, i+1)
				break
			}
		}
	}
	return idxs, nil
}

// AskPassword prompts user for a password with input masking.
// Menggantikan logika manual golang.org/x/term di terminal/input.go
// AskPassword menampilkan prompt untuk password.
func AskPassword(message string, validator survey.Validator) (string, error) {
	answer := ""
	prompt := &survey.Password{
		Message: message,
	}

	// Build options conditionally so validator can be nil (allowed).
	var opts []survey.AskOpt
	if validator != nil {
		opts = append(opts, survey.WithValidator(validator))
	}

	err := survey.AskOne(prompt, &answer, opts...)
	return answer, err
}

func AskInt(message string, defaultValue int, validator survey.Validator) (int, error) {
	var answer string
	prompt := &survey.Input{
		Message: message,
		Default: fmt.Sprintf("%d", defaultValue),
	}
	if err := survey.AskOne(prompt, &answer, survey.WithValidator(validator)); err != nil {
		return 0, err
	}
	val, _ := strconv.Atoi(answer)
	return val, nil
}

func AskString(message, defaultValue string, validator survey.Validator) (string, error) {
	var answer string
	prompt := &survey.Input{
		Message: message,
		Default: defaultValue,
	}
	err := survey.AskOne(prompt, &answer, survey.WithValidator(validator))
	return answer, err
}

func AskYesNo(question string, defaultValue bool) (bool, error) {
	var response bool
	prompt := &survey.Confirm{
		Message: question,
		Default: defaultValue,
	}
	return response, survey.AskOne(prompt, &response)
}

// sanitizeFileName membersihkan nama file dari karakter ilegal
func SanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "database"
	}

	// Ganti karakter ilegal dengan underscore
	// Daftar karakter ilegal: space, /, \, :, *, ?, ", <, >, |
	// Bisa ditambah sesuai kebutuhan
	// Referensi: https://stackoverflow.com/a/31976060
	// Note: Tidak semua karakter ilegal di semua OS, tapi ini umum.
	// Untuk validasi lebih ketat, bisa gunakan regex.
	// Misal: `[^a-zA-Z0-9._-]` untuk hanya mengizinkan alfanumerik, titik, underscore, dan dash.

	illegal := " /\\:*?\"<>|"
	name = strings.Map(func(r rune) rune {
		if strings.ContainsRune(illegal, r) {
			return '_'
		}
		return r
	}, name)

	return name
}
