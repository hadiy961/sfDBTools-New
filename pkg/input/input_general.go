// File : pkg/input/input_general.go
// Deskripsi : Fungsi utilitas untuk input interaktif dari user
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package input

import (
	"fmt"
	"os"
	"path/filepath"
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

	var opts []survey.AskOpt
	if validator != nil {
		opts = append(opts, survey.WithValidator(validator))
	}

	if err := survey.AskOne(prompt, &answer, opts...); err != nil {
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

	var opts []survey.AskOpt
	if validator != nil {
		opts = append(opts, survey.WithValidator(validator))
	}

	err := survey.AskOne(prompt, &answer, opts...)
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

// PromptString meminta input string dari user
func PromptString(message string) (string, error) {
	return AskString(message, "", nil)
}

// PromptPassword meminta input password dari user
func PromptPassword(message string) (string, error) {
	return AskPassword(message, nil)
}

// PromptConfirm meminta konfirmasi yes/no dari user
func PromptConfirm(message string) (bool, error) {
	return AskYesNo(message, true)
}

// SelectSingleFromList menampilkan list dan meminta user memilih satu item
func SelectSingleFromList(items []string, message string) (string, error) {
	var selected string
	prompt := &survey.Select{
		Message: message,
		Options: items,
	}
	err := survey.AskOne(prompt, &selected, survey.WithStdio(
		os.Stdin,
		os.Stdout,
		os.Stderr,
	))
	return selected, err
}

// SelectFileInteractive menampilkan file selector interaktif dengan fitur browse directory
// Jika user input path directory, akan tampilkan list file di directory tersebut
// Jika user input path file, akan konfirmasi yes/no
func SelectFileInteractive(directory string, message string, extensions []string) (string, error) {
	fmt.Println()
	fmt.Println("📁 File Selector - Tekan Enter untuk browse directory saat ini")
	fmt.Printf("   Directory: %s\n", directory)
	fmt.Println("   Tips: Masukkan '.' untuk browse directory saat ini")
	fmt.Println()

	for {
		var userInput string
		prompt := &survey.Input{
			Message: message,
			Default: directory,
			Help:    "Tekan Enter untuk browse, atau ketik path directory/file",
		}
		err := survey.AskOne(prompt, &userInput, survey.WithStdio(
			os.Stdin,
			os.Stdout,
			os.Stderr,
		))
		if err != nil {
			return "", err
		}

		userInput = strings.TrimSpace(userInput)
		if userInput == "" {
			userInput = directory
		}

		// Expand home directory
		if strings.HasPrefix(userInput, "~") {
			home, err := os.UserHomeDir()
			if err == nil {
				userInput = strings.Replace(userInput, "~", home, 1)
			}
		}

		// Cek apakah path ada
		fileInfo, err := os.Stat(userInput)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("Path tidak ditemukan: %s\n", userInput)
				continue
			}
			return "", fmt.Errorf("gagal mengakses path: %w", err)
		}

		// Jika directory, tampilkan list file
		if fileInfo.IsDir() {
			selectedFile, err := selectFileFromDirectory(userInput, extensions)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			if selectedFile == "" {
				// User cancel atau pilih browse lagi
				continue
			}
			return selectedFile, nil
		}

		// Jika file, konfirmasi
		confirmed, err := AskYesNo(fmt.Sprintf("Gunakan file ini: %s?", userInput), true)
		if err != nil {
			return "", err
		}
		if confirmed {
			return userInput, nil
		}
		// Jika tidak confirmed, loop lagi
	}
}

// selectFileFromDirectory menampilkan list file dalam directory dengan filter ekstensi
func selectFileFromDirectory(directory string, extensions []string) (string, error) {
	files, err := os.ReadDir(directory)
	if err != nil {
		return "", fmt.Errorf("gagal membaca directory: %w", err)
	}

	// Pisahkan directories dan files
	var subDirs []string
	var matchedFiles []string

	for _, file := range files {
		if file.IsDir() {
			subDirs = append(subDirs, file.Name()+"/")
		} else {
			fileName := file.Name()
			if matchesExtension(fileName, extensions) {
				matchedFiles = append(matchedFiles, fileName)
			}
		}
	}

	// Build options list
	var options []string

	// Tambahkan parent directory jika bukan root
	if directory != "/" {
		options = append(options, "📁 .. (parent directory)")
	}

	// Tambahkan subdirectories
	for _, dir := range subDirs {
		options = append(options, "📁 "+dir)
	}

	// Tambahkan separator jika ada files
	if len(matchedFiles) > 0 && (len(subDirs) > 0 || directory != "/") {
		options = append(options, "────────────────────────")
	}

	// Tambahkan files
	for _, file := range matchedFiles {
		options = append(options, "📄 "+file)
	}

	// Tambahkan opsi lain
	if len(options) > 0 {
		options = append(options, "────────────────────────")
	}
	options = append(options, "⌨️  [ Masukkan path manual ]")

	if len(matchedFiles) == 0 && len(subDirs) == 0 {
		fmt.Printf("\n⚠️  Tidak ada file backup atau subdirectory di: %s\n\n", directory)
	}

	var selected string
	prompt := &survey.Select{
		Message:  fmt.Sprintf("Browse: %s (Files: %d, Dirs: %d)", directory, len(matchedFiles), len(subDirs)),
		Options:  options,
		PageSize: 15,
	}
	err = survey.AskOne(prompt, &selected, survey.WithStdio(
		os.Stdin,
		os.Stdout,
		os.Stderr,
	))
	if err != nil {
		return "", err
	}

	// Handle selection
	if selected == "⌨️  [ Masukkan path manual ]" {
		return "", nil // Return empty untuk trigger loop lagi
	}

	if selected == "────────────────────────" {
		return "", nil // Separator, retry
	}

	// Remove emoji prefix
	selected = strings.TrimPrefix(selected, "📁 ")
	selected = strings.TrimPrefix(selected, "📄 ")
	selected = strings.TrimSpace(selected)

	// Handle parent directory
	if selected == ".. (parent directory)" {
		parentDir := filepath.Dir(directory)
		return selectFileFromDirectory(parentDir, extensions)
	}

	// Handle subdirectory
	if strings.HasSuffix(selected, "/") {
		subPath := filepath.Join(directory, strings.TrimSuffix(selected, "/"))
		return selectFileFromDirectory(subPath, extensions)
	}

	// Return selected file with full path
	return filepath.Join(directory, selected), nil
} // matchesExtension cek apakah filename match dengan ekstensi yang diizinkan
func matchesExtension(fileName string, extensions []string) bool {
	lowerName := strings.ToLower(fileName)
	for _, ext := range extensions {
		if strings.HasSuffix(lowerName, ext) {
			return true
		}
	}
	return false
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
