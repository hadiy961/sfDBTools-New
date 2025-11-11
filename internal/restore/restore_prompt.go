// File : internal/restore/restore_prompt.go
// Deskripsi : Interactive prompts untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package restore

import (
	"fmt"
	"regexp"
	"sfDBTools/pkg/database"
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

// promptConfirmation meminta konfirmasi user dengan pesan custom
func (s *Service) promptConfirmation(message string) (bool, error) {
	confirmed, err := input.AskYesNo(message, true)
	if err != nil {
		return false, validation.HandleInputError(err)
	}
	return confirmed, nil
}
