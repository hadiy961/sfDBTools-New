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
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/charmbracelet/lipgloss"
)

// --- Styling Dasar (Menggantikan ANSI Codes di formatting.go) ---
var (
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) // Merah
)

// PrintError prints error message (menggunakan SafePrint coordination dari terminal.go)
func PrintError(message string) {
	fmt.Fprintf(os.Stderr, "%s\n", errorStyle.Render("❌ "+message))
}

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
		if err == terminal.InterruptErr {
			return 0, fmt.Errorf("menu dibatalkan oleh pengguna")
		}
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
		if err == terminal.InterruptErr {
			return nil, fmt.Errorf("menu dibatalkan oleh pengguna")
		}
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

// AskInt menampilkan prompt untuk input integer dengan validasi.
func AskInt(message string, defaultValue int, validator survey.Validator) (int, error) {
	strDefault := fmt.Sprintf("%d", defaultValue)
	answerStr := ""
	prompt := &survey.Input{
		Message: message,
		Default: strDefault,
	}

	opts := survey.WithValidator(validator)

	err := survey.AskOne(prompt, &answerStr, opts)
	if err != nil {
		return 0, err
	}

	// Konversi final setelah validasi berhasil
	val, _ := strconv.Atoi(answerStr)
	return val, nil
}

// AskString menampilkan prompt untuk input string dengan validasi.
func AskString(message, defaultValue string, validator survey.Validator) (string, error) {
	answer := ""
	prompt := &survey.Input{
		Message: message,
		Default: defaultValue,
	}

	// Bungkus validator dalam survey.WithValidator
	opts := survey.WithValidator(validator)

	err := survey.AskOne(prompt, &answer, opts)
	return answer, err
}

// AskYesNo prompts user for yes/no input with default value
func AskYesNo(question string, defaultValue bool) (bool, error) {
	var response bool

	prompt := &survey.Confirm{
		Message: question,
		Default: defaultValue,
	}

	err := survey.AskOne(prompt, &response, survey.WithStdio(os.Stdin, os.Stdout, os.Stderr))
	if err != nil {
		if err == terminal.InterruptErr {
			return false, terminal.InterruptErr
		}
		return false, err
	}
	return response, nil
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

	// Ganti karakter ilegal dengan underscore
	replacements := map[string]string{
		" ":  "_",
		"/":  "_",
		"\\": "_",
		":":  "_",
		"*":  "_",
		"?":  "_",
		"\"": "_",
		"<":  "_",
		">":  "_",
		"|":  "_",
	}

	for old, new := range replacements {
		name = strings.ReplaceAll(name, old, new)
	}

	return name
}
