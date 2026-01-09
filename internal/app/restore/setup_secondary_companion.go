// File : internal/restore/setup_secondary_companion.go
// Deskripsi : Helper untuk companion (_dmart) pada restore secondary
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 5 Januari 2026
package restore

import (
	"fmt"
	"os"
	"path/filepath"
	backupfile "sfdbtools/internal/app/backup/helpers/file"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"strings"
)

func (s *Service) resolveSecondaryCompanionFile(allowInteractive bool) error {
	opts := s.RestoreSecondaryOpts
	if opts == nil || !opts.IncludeDmart || opts.From != "file" {
		return nil
	}

	if err := s.ensureValidCompanionPath(opts, allowInteractive); err != nil {
		return err
	}
	if strings.TrimSpace(opts.CompanionFile) != "" {
		return nil
	}

	if err := s.enforceForceWithoutDetect(opts); err != nil {
		return err
	}
	if !opts.IncludeDmart {
		return nil
	}

	if !opts.AutoDetectDmart {
		return s.handleManualCompanionSelection(opts, allowInteractive)
	}

	if s.trySetAutoDetectedCompanion(opts) {
		return nil
	}

	return s.handleCompanionNotFound(opts, allowInteractive)
}

func (s *Service) ensureValidCompanionPath(opts *restoremodel.RestoreSecondaryOptions, allowInteractive bool) error {
	if strings.TrimSpace(opts.CompanionFile) == "" {
		return nil
	}

	fi, err := os.Stat(opts.CompanionFile)
	if err == nil && !fi.IsDir() {
		return nil
	}

	if !allowInteractive {
		if opts.StopOnError {
			return fmt.Errorf("dmart file (_dmart) tidak ditemukan/invalid: %s", opts.CompanionFile)
		}
		print.PrintWarning("⚠️  Skip restore companion database (_dmart) karena dmart file invalid")
		opts.IncludeDmart = false
		opts.CompanionFile = ""
		return nil
	}

	print.PrintWarning(fmt.Sprintf("⚠️  Dmart file tidak valid: %s", opts.CompanionFile))
	opts.CompanionFile = ""
	return nil
}

func (s *Service) enforceForceWithoutDetect(opts *restoremodel.RestoreSecondaryOptions) error {
	if !opts.Force || opts.AutoDetectDmart {
		return nil
	}
	if opts.StopOnError {
		return fmt.Errorf("auto-detect dmart dimatikan (dmart-detect=false) dan mode non-interaktif (--force) aktif")
	}
	print.PrintWarning("⚠️  Skip restore companion database (_dmart) (non-interaktif, companion tidak ditentukan)")
	opts.IncludeDmart = false
	return nil
}

func (s *Service) handleManualCompanionSelection(opts *restoremodel.RestoreSecondaryOptions, allowInteractive bool) error {
	if !allowInteractive {
		return nil
	}

	confirmed, err := prompt.Confirm("Pilih file companion (_dmart) secara manual?", true)
	if err != nil || !confirmed {
		opts.IncludeDmart = false
		return nil
	}

	selectedFile, err := s.selectDmartFileInteractive(filepath.Dir(opts.File))
	if err != nil {
		return err
	}
	opts.CompanionFile = selectedFile
	return nil
}

func (s *Service) trySetAutoDetectedCompanion(opts *restoremodel.RestoreSecondaryOptions) bool {
	companionPath, err := s.detectCompanionAuto(opts.File)
	if err == nil && companionPath != "" {
		opts.CompanionFile = companionPath
		return true
	}
	return false
}

func (s *Service) handleCompanionNotFound(opts *restoremodel.RestoreSecondaryOptions, allowInteractive bool) error {
	if !allowInteractive {
		if opts.StopOnError {
			return fmt.Errorf("dmart file (_dmart) tidak ditemukan secara otomatis")
		}
		print.PrintWarning("⚠️  Skip restore companion database (_dmart) (companion tidak ditemukan)")
		opts.IncludeDmart = false
		return nil
	}

	skip, err := decideSecondaryCompanionNotFoundInteractive()
	if err != nil {
		return err
	}
	if skip {
		opts.IncludeDmart = false
		return nil
	}

	selectedFile, err := s.selectDmartFileInteractive(filepath.Dir(opts.File))
	if err != nil {
		return err
	}
	opts.CompanionFile = selectedFile
	return nil
}

func (s *Service) selectDmartFileInteractive(dir string) (string, error) {
	validExtensions := backupfile.ValidBackupFileExtensionsForSelection()
	selectedFile, err := prompt.SelectFile(dir, "Masukkan path directory atau file dmart", validExtensions)
	if err != nil {
		return "", fmt.Errorf("gagal memilih dmart file: %w", err)
	}
	return selectedFile, nil
}

func decideSecondaryCompanionNotFoundInteractive() (bool, error) {
	print.PrintWarning("⚠️  Companion (_dmart) file tidak ditemukan secara otomatis")
	selected, _, err := prompt.SelectOne("Pilih tindakan untuk companion (_dmart)", []string{
		"📁 [ Browse / pilih dmart file ]",
		"⏭️  [ Skip restore companion database (_dmart) ]",
	}, 0)
	if err != nil {
		return false, fmt.Errorf("gagal memilih opsi companion: %w", err)
	}
	return strings.HasPrefix(selected, "⏭️"), nil
}
