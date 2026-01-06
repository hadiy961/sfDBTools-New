package input

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

// SelectFileInteractive menampilkan file selector interaktif dengan fitur browse directory.
// Jika user input path directory, akan tampilkan list file di directory tersebut.
// Jika user input path file, akan konfirmasi yes/no.
func SelectFileInteractive(directory string, message string, extensions []string) (string, error) {
	if err := ensureInteractiveAllowed(); err != nil {
		return "", err
	}
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

		fileInfo, err := os.Stat(userInput)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("Path tidak ditemukan: %s\n", userInput)
				continue
			}
			return "", fmt.Errorf("gagal mengakses path: %w", err)
		}

		if fileInfo.IsDir() {
			selectedFile, err := selectFileFromDirectory(userInput, extensions)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			if selectedFile == "" {
				continue
			}
			return selectedFile, nil
		}

		confirmed, err := AskYesNo(fmt.Sprintf("Gunakan file ini: %s?", userInput), true)
		if err != nil {
			return "", err
		}
		if confirmed {
			return userInput, nil
		}
	}
}
