// File : internal/restore/setup_shared.go
// Deskripsi : Shared setup functions untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
)

// resolveBackupFile resolve lokasi file backup
func (s *Service) resolveBackupFile(filePath *string, allowInteractive bool) error {
	return s.resolveFileWithPrompt(filePath, allowInteractive,
		helper.ValidBackupFileExtensionsForSelection(),
		"file backup",
		"Masukkan path directory atau file backup")
}

// resolveSelectionCSV resolve lokasi file CSV untuk restore selection
func (s *Service) resolveSelectionCSV(csvPath *string, allowInteractive bool) error {
	return s.resolveFileWithPrompt(csvPath, allowInteractive,
		[]string{".csv"},
		"file CSV",
		"Masukkan path CSV selection")
}

// resolveFileWithPrompt adalah fungsi umum untuk resolve file dengan validasi dan prompt
func (s *Service) resolveFileWithPrompt(filePath *string, allowInteractive bool, validExtensions []string, purpose, prompt string) error {
	if strings.TrimSpace(*filePath) == "" {
		if !allowInteractive {
			return fmt.Errorf("%s wajib diisi pada mode non-interaktif (--force)", purpose)
		}

		defaultDir := s.Config.Backup.Output.BaseDirectory
		if defaultDir == "" {
			defaultDir = "."
		}

		selectedFile, err := input.SelectFileInteractive(defaultDir, prompt, validExtensions)
		if err != nil {
			return fmt.Errorf("gagal memilih %s: %w", purpose, err)
		}
		*filePath = selectedFile
	}

	// Validasi file
	if err := validateFileExists(*filePath); err != nil {
		return fmt.Errorf("%s: %w", purpose, err)
	}

	fi, err := os.Stat(*filePath)
	if err != nil {
		return fmt.Errorf("gagal membaca %s: %w", purpose, err)
	}

	if fi.IsDir() {
		if !allowInteractive {
			return fmt.Errorf("%s tidak valid (path adalah direktori): %s", purpose, *filePath)
		}

		selectedFile, selErr := input.SelectFileInteractive(*filePath, prompt, validExtensions)
		if selErr != nil {
			return fmt.Errorf("gagal memilih %s: %w", purpose, selErr)
		}
		*filePath = selectedFile
	}

	if err := validateFileExtension(*filePath, validExtensions, purpose); err != nil {
		if !allowInteractive {
			return err
		}
		ui.PrintWarning(fmt.Sprintf("⚠️  %s", err.Error()))
		selectedFile, selErr := input.SelectFileInteractive(filepath.Dir(*filePath), prompt, validExtensions)
		if selErr != nil {
			return fmt.Errorf("gagal memilih %s: %w", purpose, selErr)
		}
		*filePath = selectedFile
	}

	absPath, err := filepath.Abs(*filePath)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan absolute path: %w", err)
	}
	*filePath = absPath

	s.Log.Infof("%s: %s", strings.Title(purpose), *filePath)
	return nil
}

// resolveEncryptionKey resolve encryption key untuk decrypt file
func (s *Service) resolveEncryptionKey(filePath string, encryptionKey *string, allowInteractive bool) error {
	if !helper.IsEncryptedFile(filePath) {
		s.Log.Debug("File backup tidak terenkripsi")
		return nil
	}

	return s.validateAndRetryEncryptionKey(filePath, encryptionKey, allowInteractive)
}
