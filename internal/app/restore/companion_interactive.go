// File : internal/restore/companion_interactive.go
// Deskripsi : Helper interaktif untuk pemilihan companion (_dmart) file
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package restore

import (
	"fmt"
	"path/filepath"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"strings"
)

func (s *Service) useOrSelectDetectedCompanionInteractive(detectedPath string) error {
	if s.RestorePrimaryOpts.Force {
		s.RestorePrimaryOpts.CompanionFile = detectedPath
		return nil
	}

	options := []string{
		fmt.Sprintf("✅ [ Pakai dmart file terdeteksi: %s ]", filepath.Base(detectedPath)),
		"📁 [ Browse / pilih dmart file lain ]",
		"⏭️  [ Skip restore companion database (_dmart) ]",
	}
	selected, err := input.SelectSingleFromList(options, "Companion (dmart) file ditemukan. Gunakan file ini atau pilih yang lain?")
	if err == nil && selected == options[0] {
		s.RestorePrimaryOpts.CompanionFile = detectedPath
		s.Log.Infof("Companion file dipakai (auto-detect): %s", filepath.Base(detectedPath))
		return nil
	}
	if err == nil && selected == options[2] {
		s.RestorePrimaryOpts.CompanionFile = ""
		s.RestorePrimaryOpts.IncludeDmart = false
		s.Log.Info("Companion database tidak akan di-restore (user memilih skip)")
		ui.PrintWarning("⚠️  Skip restore companion database (_dmart)")
		return nil
	}

	chosen, bErr := s.browseCompanionFileInteractive()
	if bErr != nil {
		return bErr
	}
	s.RestorePrimaryOpts.CompanionFile = chosen
	return nil
}

func (s *Service) selectCompanionFileInteractive() error {
	confirm, err := input.PromptConfirm("Apakah Anda ingin memilih file companion database (_dmart) secara manual?")
	if err != nil {
		return fmt.Errorf("gagal mendapatkan konfirmasi: %w", err)
	}

	if !confirm {
		s.Log.Info("User memilih untuk skip restore companion database")
		ui.PrintWarning("⚠️  Skip restore companion database")
		s.RestorePrimaryOpts.IncludeDmart = false
		return nil
	}

	chosen, err := s.browseCompanionFileInteractive()
	if err != nil {
		return err
	}
	s.RestorePrimaryOpts.CompanionFile = chosen
	s.Log.Infof("User memilih companion file: %s", filepath.Base(chosen))

	return nil
}

func (s *Service) browseCompanionFileInteractive() (string, error) {
	dir := filepath.Dir(s.RestorePrimaryOpts.File)
	files, err := helper.ListBackupFilesInDirectory(dir)
	if err != nil {
		return "", fmt.Errorf("gagal list files: %w", err)
	}
	if len(files) == 0 {
		return "", fmt.Errorf("tidak ada file backup ditemukan di direktori: %s", dir)
	}

	choice, err := input.ShowMenu("Pilih file companion database:", files)
	if err != nil {
		return "", fmt.Errorf("gagal memilih file: %w", err)
	}

	selected := strings.TrimSpace(files[choice-1])
	return filepath.Join(dir, selected), nil
}
