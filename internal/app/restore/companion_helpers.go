// File : internal/restore/companion_helpers.go
// Deskripsi : Helper functions untuk companion database detection
// Author : Hadiyatna Muflihun
// Tanggal : 19 Desember 2025
// Last Modified : 30 Desember 2025

package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
	"strings"
)

// DetectOrSelectCompanionFile mendeteksi atau meminta user memilih file companion database
func (s *Service) DetectOrSelectCompanionFile() error {
	if s.RestorePrimaryOpts == nil {
		return fmt.Errorf("restore primary options tidak tersedia")
	}

	handled, err := s.useCompanionFileFromFlagOrDecide()
	if err != nil {
		return err
	}
	if handled {
		return nil
	}

	// Non-interactive mode: jangan pernah prompt.
	// Jika companion tidak bisa ditentukan otomatis, maka:
	// - StopOnError=true  => return error
	// - StopOnError=false => skip companion restore (IncludeDmart=false) dan lanjut
	opts := s.RestorePrimaryOpts
	if opts.Force && !opts.AutoDetectDmart {
		err := fmt.Errorf("auto-detect dmart dimatikan (dmart-detect=false) dan mode non-interaktif (--force) aktif; tentukan dmart via --dmart-file atau set --dmart-include=false")
		return s.stopOrSkipCompanionNonInteractive(
			err,
			"Auto-detect companion dimatikan dan mode non-interaktif aktif; skip restore companion database",
			"⚠️  Skip restore companion database (non-interaktif, companion tidak ditentukan)",
		)
	}

	if !opts.AutoDetectDmart {
		return s.selectCompanionFileInteractive()
	}

	primaryFile := opts.File
	companionPath, err := s.detectCompanionAuto(primaryFile)
	if err == nil && companionPath != "" {
		s.Log.Infof("✓ Companion file ditemukan: %s", filepath.Base(companionPath))
		return s.useOrSelectDetectedCompanionInteractive(companionPath)
	}
	s.Log.Debugf("Auto-detect companion gagal: %v", err)

	// Not found, ask user
	ui.PrintWarning("⚠️  Companion file tidak ditemukan secara otomatis")
	if opts.Force {
		err := fmt.Errorf("dmart file (_dmart) tidak ditemukan secara otomatis dan mode non-interaktif (--force) aktif; gunakan --dmart-file untuk set file dmart, atau set --dmart-include=false, atau gunakan --continue-on-error untuk skip dmart")
		return s.stopOrSkipCompanionNonInteractive(
			err,
			"Companion file tidak ditemukan; skip restore companion database karena continue-on-error",
			"⚠️  Skip restore companion database (companion tidak ditemukan)",
		)
	}
	return s.selectCompanionFileInteractive()
}

func (s *Service) useCompanionFileFromFlagOrDecide() (bool, error) {
	opts := s.RestorePrimaryOpts
	flagPath := strings.TrimSpace(opts.CompanionFile)
	if flagPath == "" {
		return false, nil
	}
	opts.CompanionFile = flagPath

	fi, err := os.Stat(flagPath)
	if err != nil {
		s.Log.Warnf("Companion file dari flag tidak ditemukan: %s", flagPath)
		ui.PrintWarning(fmt.Sprintf("⚠️  Companion file tidak ditemukan: %s", flagPath))
		if opts.Force {
			return true, s.stopOrSkipCompanionNonInteractive(
				fmt.Errorf("companion file (_dmart) tidak ditemukan: %s", flagPath),
				"Companion file tidak ditemukan; skip restore companion database karena continue-on-error",
				"⚠️  Skip restore companion database (companion file tidak ditemukan)",
			)
		}
		opts.CompanionFile = ""
		return false, nil
	}

	if fi.IsDir() {
		s.Log.Warnf("Companion file dari flag adalah direktori (tidak valid): %s", flagPath)
		ui.PrintWarning(fmt.Sprintf("⚠️  Companion file tidak valid (path adalah direktori): %s", flagPath))
		if opts.Force {
			return true, s.stopOrSkipCompanionNonInteractive(
				fmt.Errorf("companion file (_dmart) tidak valid (path adalah direktori): %s", flagPath),
				"Companion file tidak valid; skip restore companion database karena continue-on-error",
				"⚠️  Skip restore companion database (companion file tidak valid)",
			)
		}
		opts.CompanionFile = ""
		return false, nil
	}

	validExtensions := helper.ValidBackupFileExtensionsForSelection()
	if !isValidBackupFileExtension(flagPath, validExtensions) {
		s.Log.Warnf("Companion file dari flag tidak valid (ekstensi): %s", flagPath)
		ui.PrintWarning(fmt.Sprintf("⚠️  Companion file tidak valid (ekstensi tidak didukung): %s", flagPath))
		if opts.Force {
			return true, s.stopOrSkipCompanionNonInteractive(
				fmt.Errorf("companion file (_dmart) tidak valid (ekstensi tidak didukung): %s", flagPath),
				"Companion file tidak valid; skip restore companion database karena continue-on-error",
				"⚠️  Skip restore companion database (companion file tidak valid)",
			)
		}
		opts.CompanionFile = ""
		return false, nil
	}

	s.Log.Infof("Menggunakan companion file yang sudah ditentukan: %s", flagPath)
	return true, nil
}

func isValidBackupFileExtension(path string, validExtensions []string) bool {
	lower := strings.ToLower(strings.TrimSpace(path))
	for _, ext := range validExtensions {
		if strings.HasSuffix(lower, strings.ToLower(ext)) {
			return true
		}
	}
	return false
}

func (s *Service) stopOrSkipCompanionNonInteractive(err error, warnLog string, warnUI string) error {
	opts := s.RestorePrimaryOpts
	if !opts.Force {
		return err
	}
	if opts.StopOnError {
		return err
	}
	s.skipCompanionRestore(warnLog, warnUI)
	return nil
}

func (s *Service) skipCompanionRestore(warnLog string, warnUI string) {
	s.Log.Warn(warnLog)
	ui.PrintWarning(warnUI)
	s.RestorePrimaryOpts.IncludeDmart = false
	s.RestorePrimaryOpts.CompanionFile = ""
}
