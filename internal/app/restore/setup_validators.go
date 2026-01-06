// File : internal/restore/setup_validators.go
// Deskripsi : Validation functions untuk setup restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 6 Januari 2026
package restore

import (
	"fmt"
	"io"
	"os"
	"sfdbtools/internal/app/restore/helpers"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/pkg/consts"
	"strings"
)

// validateFileExists memvalidasi keberadaan file
func validateFileExists(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file tidak ditemukan: %s", filePath)
	} else if err != nil {
		return fmt.Errorf("gagal membaca file: %w", err)
	}
	return nil
}

// validateFileExtension memvalidasi ekstensi file
func validateFileExtension(filePath string, validExtensions []string, purpose string) error {
	if !hasAnySuffix(filePath, validExtensions) {
		return fmt.Errorf("%s ekstensi tidak didukung: %s", purpose, filePath)
	}
	return nil
}

// validateEncryptionKey memvalidasi encryption key terhadap file
func validateEncryptionKey(filePath string, key string) error {
	reader, closers, err := helpers.OpenAndPrepareReader(filePath, key)
	if err != nil {
		return fmt.Errorf("gagal membuka file dengan key: %w", err)
	}
	defer helpers.CloseReaders(closers)

	buf := make([]byte, 1)
	_, readErr := reader.Read(buf)
	if readErr != nil && readErr != io.EOF {
		return fmt.Errorf("gagal membaca file: %w", readErr)
	}

	return nil
}

// validateAndRetryEncryptionKey memvalidasi encryption key dengan retry interaktif
func (s *Service) validateAndRetryEncryptionKey(filePath string, encryptionKey *string, allowInteractive bool) error {
	for {
		if strings.TrimSpace(*encryptionKey) == "" {
			if !allowInteractive {
				return fmt.Errorf("file backup terenkripsi; encryption key wajib diisi (--enc-key atau env) pada mode non-interaktif (--force)")
			}
			key, err := prompt.PromptPassword("Masukkan encryption key untuk decrypt file backup")
			if err != nil {
				return fmt.Errorf("gagal mendapatkan encryption key: %w", err)
			}
			*encryptionKey = key
		}

		if err := validateEncryptionKey(filePath, *encryptionKey); err == nil {
			return nil
		} else {
			if !allowInteractive {
				return fmt.Errorf("validasi encryption key gagal: %w", err)
			}

			print.PrintError(fmt.Sprintf("Encryption key tidak valid atau file gagal didecrypt: %v", err))
			action, _, selErr := prompt.SelectOne(
				"Encryption key salah. Pilih aksi:",
				[]string{"Ubah key enkripsi", "Batalkan"},
				0,
			)
			if selErr != nil {
				return fmt.Errorf("gagal memilih aksi setelah error encryption key: %w", selErr)
			}

			if action == "Batalkan" {
				return fmt.Errorf("restore dibatalkan oleh user (encryption key salah)")
			}

			*encryptionKey = ""
		}
	}
}

// validateClientCodeInput memvalidasi input client code
func validateClientCodeInput(ans interface{}) error {
	str, ok := ans.(string)
	if !ok {
		return fmt.Errorf("input tidak valid")
	}
	if strings.TrimSpace(str) == "" {
		return fmt.Errorf("client-code tidak boleh kosong")
	}
	return nil
}

// validatePrimaryClientCodeInput memvalidasi input client code untuk primary database
func validatePrimaryClientCodeInput(prefix string) func(interface{}) error {
	return func(ans interface{}) error {
		str, ok := ans.(string)
		if !ok {
			return fmt.Errorf("input tidak valid")
		}
		str = strings.TrimSpace(str)
		if str == "" {
			return fmt.Errorf("client-code tidak boleh kosong")
		}

		candidate := buildPrimaryTargetDBFromClientCode(prefix, str)
		if !helpers.IsPrimaryDatabaseName(candidate) {
			return fmt.Errorf("client-code menghasilkan nama database primary yang tidak valid")
		}
		return nil
	}
}

// validateGrantsFilePath memvalidasi path file grants
func validateGrantsFilePath(filePath string, stopOnError bool) error {
	fi, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		if stopOnError {
			return fmt.Errorf("file grants tidak ditemukan: %s", filePath)
		}
		return nil
	}
	if err != nil {
		if stopOnError {
			return fmt.Errorf("gagal membaca file grants: %w", err)
		}
		return nil
	}
	if fi.IsDir() {
		if stopOnError {
			return fmt.Errorf("file grants tidak valid (path adalah direktori): %s", filePath)
		}
		return nil
	}

	if !strings.HasSuffix(strings.ToLower(filePath), strings.ToLower(consts.UsersSQLSuffix)) {
		if stopOnError {
			return fmt.Errorf("file grants tidak valid (harus berakhiran '%s'): %s", consts.UsersSQLSuffix, filePath)
		}
		return nil
	}

	return nil
}

// validateCompanionFile memvalidasi file companion (dmart)
func (s *Service) validateCompanionFile(opts interface{}, allowInteractive bool) error {
	// Type assertion untuk mendapatkan CompanionFile
	var companionFile string
	var stopOnError bool
	var includeDmart *bool

	switch v := opts.(type) {
	case *restoremodel.RestoreSecondaryOptions:
		companionFile = v.CompanionFile
		stopOnError = v.StopOnError
		includeDmart = &v.IncludeDmart
	case *restoremodel.RestorePrimaryOptions:
		companionFile = v.CompanionFile
		stopOnError = v.StopOnError
		includeDmart = &v.IncludeDmart
	default:
		return fmt.Errorf("tipe options tidak didukung")
	}

	if strings.TrimSpace(companionFile) == "" {
		return nil
	}

	fi, err := os.Stat(companionFile)
	if err == nil && !fi.IsDir() {
		return nil
	}

	if !allowInteractive {
		if stopOnError {
			return fmt.Errorf("dmart file (_dmart) tidak ditemukan/invalid: %s", companionFile)
		}
		print.PrintWarning("⚠️  Skip restore companion database (_dmart) karena dmart file invalid")
		*includeDmart = false
		return nil
	}

	print.PrintWarning(fmt.Sprintf("⚠️  Dmart file tidak valid: %s", companionFile))
	return nil
}
