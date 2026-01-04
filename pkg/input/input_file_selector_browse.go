package input

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

// selectFileFromDirectory menampilkan list file dalam directory dengan filter ekstensi.
func selectFileFromDirectory(directory string, extensions []string) (string, error) {
	if err := ensureInteractiveAllowed(); err != nil {
		return "", err
	}
	files, err := os.ReadDir(directory)
	if err != nil {
		return "", fmt.Errorf("gagal membaca directory: %w", err)
	}

	var subDirs []string
	var matchedFiles []string

	for _, file := range files {
		if file.IsDir() {
			subDirs = append(subDirs, file.Name()+"/")
			continue
		}

		fileName := file.Name()
		if matchesExtension(fileName, extensions) {
			matchedFiles = append(matchedFiles, fileName)
		}
	}

	var options []string
	if directory != "/" {
		options = append(options, "📁 .. (parent directory)")
	}
	for _, dir := range subDirs {
		options = append(options, "📁 "+dir)
	}
	if len(matchedFiles) > 0 && (len(subDirs) > 0 || directory != "/") {
		options = append(options, "────────────────────────")
	}
	for _, file := range matchedFiles {
		options = append(options, "📄 "+file)
	}
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

	if selected == "⌨️  [ Masukkan path manual ]" {
		return "", nil
	}
	if selected == "────────────────────────" {
		return "", nil
	}

	selected = strings.TrimPrefix(selected, "📁 ")
	selected = strings.TrimPrefix(selected, "📄 ")
	selected = strings.TrimSpace(selected)

	if selected == ".. (parent directory)" {
		parentDir := filepath.Dir(directory)
		return selectFileFromDirectory(parentDir, extensions)
	}
	if strings.HasSuffix(selected, "/") {
		subPath := filepath.Join(directory, strings.TrimSuffix(selected, "/"))
		return selectFileFromDirectory(subPath, extensions)
	}

	return filepath.Join(directory, selected), nil
}
