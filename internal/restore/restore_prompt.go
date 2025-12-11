// File : internal/restore/restore_prompt.go
// Deskripsi : Interactive prompts untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package restore

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"

	"github.com/AlecAivazis/survey/v2"
)

// promptDatabaseName meminta user untuk input database name interactively
// Digunakan sebagai fallback ketika pattern extraction gagal
func (s *Service) promptDatabaseName(sourceFile string) (string, error) {
	ui.PrintSubHeader("Database Name Required")

	s.Log.Warn("Tidak dapat mendeteksi nama database dari filename")
	s.Log.Infof("Source file: %s", sourceFile)
	s.Log.Info("Silakan masukkan nama database untuk restore:")

	fmt.Println() // Spacing

	// Validator untuk database name (alphanumeric, underscore, dash)
	dbNameValidator := input.ComposeValidators(
		survey.Required,
		validateDatabaseName,
	)

	// Loop sampai user input valid database name atau cancel
	for {
		dbName, err := input.AskString(
			"Database Name",
			"", // No default value
			dbNameValidator,
		)

		if err != nil {
			return "", validation.HandleInputError(err)
		}

		// Additional validation: tidak boleh system database
		if database.IsSystemDatabase(dbName) {
			ui.PrintError(fmt.Sprintf("'%s' adalah system database, tidak bisa digunakan untuk restore", dbName))
			continue
		}

		// Confirm dengan user
		confirm, err := input.AskYesNo(
			fmt.Sprintf("Restore ke database '%s'?", dbName),
			true,
		)

		if err != nil {
			return "", validation.HandleInputError(err)
		}

		if confirm {
			return dbName, nil
		}

		// User tidak confirm, loop lagi
		fmt.Println()
	}
}

// PromptSourceFile meminta user untuk input lokasi file backup interactively
// Mendukung relative dan absolute path
func (s *Service) PromptSourceFile() (string, error) {
	ui.PrintSubHeader("Source File Required")

	s.Log.Info("Silakan masukkan lokasi file backup untuk restore:")
	s.Log.Info("(Mendukung relative dan absolute path)")

	fmt.Println() // Spacing

	// Validator untuk file path
	filePathValidator := input.ComposeValidators(
		survey.Required,
		validateFilePath,
	)

	// Loop sampai user input valid file path atau cancel
	for {
		filePath, err := input.AskString(
			"Lokasi File Backup",
			"", // No default value
			filePathValidator,
		)

		if err != nil {
			return "", validation.HandleInputError(err)
		}

		// Expand tilde dan resolve ke absolute path
		filePath = helper.ExpandPath(filePath)
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			ui.PrintError(fmt.Sprintf("Gagal resolve path: %v", err))
			continue
		}

		// Confirm dengan user (show absolute path)
		confirm, err := input.AskYesNo(
			fmt.Sprintf("Gunakan file '%s'?", absPath),
			true,
		)

		if err != nil {
			return "", validation.HandleInputError(err)
		}

		if confirm {
			return absPath, nil
		}

		// User tidak confirm, loop lagi
		fmt.Println()
	}
}

// validateDatabaseName validates database name format
// Database name harus alphanumeric dengan underscore/dash
func validateDatabaseName(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("database name harus berupa string")
	}

	if str == "" {
		return fmt.Errorf("database name tidak boleh kosong")
	}

	// Database name format: alphanumeric, underscore, dash
	// Must start with letter or underscore
	// Length: 1-64 characters (MySQL limit)
	pattern := `^[a-zA-Z_][a-zA-Z0-9_-]{0,63}$`
	matched, err := regexp.MatchString(pattern, str)

	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !matched {
		return fmt.Errorf("database name tidak valid: harus dimulai dengan huruf/underscore, hanya boleh berisi alphanumeric, underscore, dan dash (max 64 karakter)")
	}

	return nil
}

// validateFilePath validates file path format and existence
func validateFilePath(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("file path harus berupa string")
	}

	if str == "" {
		return fmt.Errorf("file path tidak boleh kosong")
	}

	// Tidak perlu check existence di sini karena akan di-check di verify
	// Cukup pastikan format tidak aneh
	return nil
}

// promptConfirmation meminta konfirmasi user dengan pesan custom
func (s *Service) promptConfirmation(message string) (bool, error) {
	confirmed, err := input.AskYesNo(message, true)
	if err != nil {
		return false, validation.HandleInputError(err)
	}
	return confirmed, nil
}
