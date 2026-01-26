// File : internal/restore/companion_helpers.go
// Deskripsi : Helper functions untuk companion database detection
// Author : Hadiyatna Muflihun
// Tanggal : 19 Desember 2025
// Last Modified : 26 Januari 2026

package restore

import (
	"fmt"
	"os"
	"path/filepath"
	backupfile "sfdbtools/internal/app/backup/helpers/file"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/ui/print"
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
	nonInteractive := opts.Force || runtimecfg.IsQuiet()
	if nonInteractive && !opts.AutoDetectDmart {
		err := fmt.Errorf("auto-detect dmart dimatikan (dmart-detect=false) dan mode non-interaktif (--skip-confirm/--quiet) aktif; tentukan dmart via --dmart-file atau set --dmart-include=false")
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
	print.PrintWarning("⚠️  Companion file tidak ditemukan secara otomatis")
	if nonInteractive {
		err := fmt.Errorf("dmart file (_dmart) tidak ditemukan secara otomatis dan mode non-interaktif (--skip-confirm/--quiet) aktif; gunakan --dmart-file untuk set file dmart, atau set --dmart-include=false, atau gunakan --continue-on-error untuk skip dmart")
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
		print.PrintWarning(fmt.Sprintf("⚠️  Companion file tidak ditemukan: %s", flagPath))
		nonInteractive := opts.Force || runtimecfg.IsQuiet()
		if nonInteractive {
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
		print.PrintWarning(fmt.Sprintf("⚠️  Companion file tidak valid (path adalah direktori): %s", flagPath))
		nonInteractive := opts.Force || runtimecfg.IsQuiet()
		if nonInteractive {
			return true, s.stopOrSkipCompanionNonInteractive(
				fmt.Errorf("companion file (_dmart) tidak valid (path adalah direktori): %s", flagPath),
				"Companion file tidak valid; skip restore companion database karena continue-on-error",
				"⚠️  Skip restore companion database (companion file tidak valid)",
			)
		}
		opts.CompanionFile = ""
		return false, nil
	}

	validExtensions := backupfile.ValidBackupFileExtensionsForSelection()
	if !isValidBackupFileExtension(flagPath, validExtensions) {
		s.Log.Warnf("Companion file dari flag tidak valid (ekstensi): %s", flagPath)
		print.PrintWarning(fmt.Sprintf("⚠️  Companion file tidak valid (ekstensi tidak didukung): %s", flagPath))
		nonInteractive := opts.Force || runtimecfg.IsQuiet()
		if nonInteractive {
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
	nonInteractive := opts.Force || runtimecfg.IsQuiet()
	if !nonInteractive {
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
	print.PrintWarning(warnUI)
	s.RestorePrimaryOpts.IncludeDmart = false
	s.RestorePrimaryOpts.CompanionFile = ""
}
