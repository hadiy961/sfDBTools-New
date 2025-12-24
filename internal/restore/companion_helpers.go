// File : internal/restore/companion_helpers.go
// Deskripsi : Helper functions untuk companion database detection
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-19
// Last Modified : 2025-12-19

package restore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"strings"
)

// DetectOrSelectCompanionFile mendeteksi atau meminta user memilih file companion database
func (s *Service) DetectOrSelectCompanionFile() error {
	// Jika companion file sudah di-set, gunakan dulu.
	// Jika ternyata file tidak ada (mis. flag salah), fallback ke auto-detect / pemilihan interaktif.
	if s.RestorePrimaryOpts.CompanionFile != "" {
		validExtensions := helper.ValidBackupFileExtensionsForSelection()
		if fi, err := os.Stat(s.RestorePrimaryOpts.CompanionFile); err == nil {
			if fi.IsDir() {
				invalid := s.RestorePrimaryOpts.CompanionFile
				s.Log.Warnf("Companion file dari flag adalah direktori (tidak valid): %s", invalid)
				ui.PrintWarning(fmt.Sprintf("⚠️  Companion file tidak valid (path adalah direktori): %s", invalid))

				// Mode non-interaktif: jangan prompt, ikuti StopOnError.
				if s.RestorePrimaryOpts.Force {
					if s.RestorePrimaryOpts.StopOnError {
						return fmt.Errorf("companion file (_dmart) tidak valid (path adalah direktori): %s", invalid)
					}
					s.Log.Warn("Companion file tidak valid; skip restore companion database karena continue-on-error")
					ui.PrintWarning("⚠️  Skip restore companion database (companion file tidak valid)")
					s.RestorePrimaryOpts.IncludeDmart = false
					return nil
				}

				// Interactive fallback
				s.RestorePrimaryOpts.CompanionFile = ""
			} else {
				// Validasi ekstensi file dmart
				lower := strings.ToLower(strings.TrimSpace(s.RestorePrimaryOpts.CompanionFile))
				valid := false
				for _, ext := range validExtensions {
					if strings.HasSuffix(lower, strings.ToLower(ext)) {
						valid = true
						break
					}
				}
				if !valid {
					invalid := s.RestorePrimaryOpts.CompanionFile
					s.Log.Warnf("Companion file dari flag tidak valid (ekstensi): %s", invalid)
					ui.PrintWarning(fmt.Sprintf("⚠️  Companion file tidak valid (ekstensi tidak didukung): %s", invalid))

					if s.RestorePrimaryOpts.Force {
						if s.RestorePrimaryOpts.StopOnError {
							return fmt.Errorf("companion file (_dmart) tidak valid (ekstensi tidak didukung): %s", invalid)
						}
						s.Log.Warn("Companion file tidak valid; skip restore companion database karena continue-on-error")
						ui.PrintWarning("⚠️  Skip restore companion database (companion file tidak valid)")
						s.RestorePrimaryOpts.IncludeDmart = false
						return nil
					}

					// Interactive fallback
					s.RestorePrimaryOpts.CompanionFile = ""
				} else {
					s.Log.Infof("Menggunakan companion file yang sudah ditentukan: %s", s.RestorePrimaryOpts.CompanionFile)
					return nil
				}
			}
		}

		missing := s.RestorePrimaryOpts.CompanionFile
		s.Log.Warnf("Companion file dari flag tidak ditemukan: %s", missing)
		ui.PrintWarning(fmt.Sprintf("⚠️  Companion file tidak ditemukan: %s", missing))

		// Mode non-interaktif: jangan prompt, ikuti StopOnError.
		if s.RestorePrimaryOpts.Force {
			if s.RestorePrimaryOpts.StopOnError {
				return fmt.Errorf("companion file (_dmart) tidak ditemukan: %s", missing)
			}
			s.Log.Warn("Companion file tidak ditemukan; skip restore companion database karena continue-on-error")
			ui.PrintWarning("⚠️  Skip restore companion database (companion file tidak ditemukan)")
			s.RestorePrimaryOpts.IncludeDmart = false
			return nil
		}

		// Interactive fallback
		s.RestorePrimaryOpts.CompanionFile = ""
	}

	// Non-interactive mode: jangan pernah prompt.
	// Jika companion tidak bisa ditentukan otomatis, maka:
	// - StopOnError=true  => return error
	// - StopOnError=false => skip companion restore (IncludeDmart=false) dan lanjut
	if s.RestorePrimaryOpts.Force && !s.RestorePrimaryOpts.AutoDetectDmart {
		if s.RestorePrimaryOpts.StopOnError {
			return fmt.Errorf("auto-detect dmart dimatikan (dmart-detect=false) dan mode non-interaktif (--force) aktif; tentukan dmart via --dmart-file atau set --dmart-include=false")
		}
		s.Log.Warn("Auto-detect companion dimatikan dan mode non-interaktif aktif; skip restore companion database")
		ui.PrintWarning("⚠️  Skip restore companion database (non-interaktif, companion tidak ditentukan)")
		s.RestorePrimaryOpts.IncludeDmart = false
		return nil
	}

	if !s.RestorePrimaryOpts.AutoDetectDmart {
		return s.selectCompanionFileInteractive()
	}

	// Auto-detect companion file
	primaryFile := s.RestorePrimaryOpts.File
	dir := filepath.Dir(primaryFile)

	s.Log.Infof("Auto-detect companion (_dmart) rule: 1) baca metadata '%s', 2) jika gagal, pattern standar, 3) fallback sibling filename", filepath.Base(primaryFile)+consts.ExtMetaJSON)
	s.Log.Info("Pattern standar: {database}_{YYYYMMDD}_{HHMMSS}_{hostname}.sql[.gz][.enc] (dan companion {database}_dmart...) ")
	s.Log.Debugf("Mencari companion file dari primary: %s", filepath.Base(primaryFile))

	// Strategi 1: Coba baca dari metadata file (.meta.json)
	companionPath, err := s.detectCompanionFromMetadata(primaryFile)
	if err == nil && companionPath != "" {
		s.Log.Infof("✓ Companion file ditemukan dari metadata: %s", filepath.Base(companionPath))
		return s.useOrSelectDetectedCompanionInteractive(companionPath)
	}
	s.Log.Debugf("Gagal detect dari metadata: %v, mencoba pattern matching", err)

	// Strategi 2: Pattern matching - cari file dengan pattern yang sesuai
	companionPath, err = s.detectCompanionByPattern(primaryFile, dir)
	if err == nil && companionPath != "" {
		s.Log.Infof("✓ Companion file ditemukan via pattern: %s", filepath.Base(companionPath))
		return s.useOrSelectDetectedCompanionInteractive(companionPath)
	}
	s.Log.Debugf("Gagal detect via pattern: %v", err)

	// Strategi 3: Fallback sederhana - cari sibling file dengan nama "{primary}_dmart" (tanpa timestamp/hostname)
	companionPath, err = s.detectCompanionBySiblingFilename(primaryFile, dir)
	if err == nil && companionPath != "" {
		s.Log.Infof("✓ Companion file ditemukan via sibling filename: %s", filepath.Base(companionPath))
		return s.useOrSelectDetectedCompanionInteractive(companionPath)
	}
	s.Log.Debugf("Gagal detect via sibling filename: %v", err)

	// Not found, ask user
	ui.PrintWarning("⚠️  Companion file tidak ditemukan secara otomatis")
	if s.RestorePrimaryOpts.Force {
		if s.RestorePrimaryOpts.StopOnError {
			return fmt.Errorf("dmart file (_dmart) tidak ditemukan secara otomatis dan mode non-interaktif (--force) aktif; gunakan --dmart-file untuk set file dmart, atau set --dmart-include=false, atau gunakan --continue-on-error untuk skip dmart")
		}
		s.Log.Warn("Companion file tidak ditemukan; skip restore companion database karena continue-on-error")
		ui.PrintWarning("⚠️  Skip restore companion database (companion tidak ditemukan)")
		s.RestorePrimaryOpts.IncludeDmart = false
		return nil
	}
	return s.selectCompanionFileInteractive()
}

func (s *Service) useOrSelectDetectedCompanionInteractive(detectedPath string) error {
	// Non-interaktif: langsung pakai file terdeteksi.
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

	// User memilih browse / atau prompt gagal: lanjut ke browse langsung.
	chosen, bErr := s.browseCompanionFileInteractive()
	if bErr != nil {
		return bErr
	}
	s.RestorePrimaryOpts.CompanionFile = chosen
	return nil
}

// detectCompanionBySiblingFilename mencoba menemukan companion file dengan pola sederhana:
//
//	primary:   dbsf_nbc_xxx[_nodata].sql(.gz/.zst/.enc)
//	companion: dbsf_nbc_xxx_dmart[_nodata].sql(.gz/.zst/.enc)
//
// Ini berguna untuk dump yang tidak memakai format timestamp/hostname dan tidak punya .meta.json.
func (s *Service) detectCompanionBySiblingFilename(primaryFile string, dir string) (string, error) {
	basename := filepath.Base(primaryFile)
	nameWithoutExt, extensions := helper.ExtractFileExtensions(basename)
	if nameWithoutExt == "" {
		return "", fmt.Errorf("gagal parse filename: %s", basename)
	}

	// nameWithoutExt di sini adalah token {database} pada filename.
	dbName := nameWithoutExt

	// Build companion database name.
	// Jika backup dibuat dengan excludeData, generator akan menambahkan suffix "_nodata".
	// Untuk companion, suffix _dmart harus berada sebelum _nodata.
	companionDBName := dbName + consts.SuffixDmart
	lowerDBName := strings.ToLower(dbName)
	if strings.HasSuffix(lowerDBName, "_nodata") {
		base := dbName[:len(dbName)-len("_nodata")]
		companionDBName = base + consts.SuffixDmart + "_nodata"
	}

	extStr := strings.Join(extensions, "")
	candidate := filepath.Join(dir, companionDBName+extStr)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	return "", fmt.Errorf("companion sibling tidak ditemukan: %s", filepath.Base(candidate))
}

// detectCompanionFromMetadata mencoba mendapatkan companion file dari metadata
func (s *Service) detectCompanionFromMetadata(primaryFile string) (string, error) {
	metadataPath := primaryFile + consts.ExtMetaJSON

	// Check if metadata exists
	if _, err := os.Stat(metadataPath); err != nil {
		return "", fmt.Errorf("metadata file tidak ditemukan: %w", err)
	}

	// Read metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return "", fmt.Errorf("gagal baca metadata: %w", err)
	}

	// Parse JSON
	var meta types_backup.BackupMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return "", fmt.Errorf("gagal parse metadata: %w", err)
	}

	// Cari companion database di DatabaseDetails
	if len(meta.DatabaseDetails) > 0 {
		for _, detail := range meta.DatabaseDetails {
			// Cari yang mengandung companion suffix di nama database
			if strings.Contains(strings.ToLower(detail.DatabaseName), consts.SuffixDmart) {
				// Validasi file exists
				if _, err := os.Stat(detail.BackupFile); err == nil {
					s.Log.Debugf("Found companion in metadata: %s", detail.DatabaseName)
					return detail.BackupFile, nil
				}
				s.Log.Warnf("Companion file di metadata tidak ada di disk: %s", detail.BackupFile)
			}
		}
	}

	return "", fmt.Errorf("tidak ada companion database ditemukan di metadata")
}

// detectCompanionByPattern mencoba menemukan companion file menggunakan pattern matching
// Format backup file: {database}_{YYYYMMDD}_{HHMMSS}_{hostname}.sql.gz[.enc]
func (s *Service) detectCompanionByPattern(primaryFile string, dir string) (string, error) {
	basename := filepath.Base(primaryFile)

	// Extract extensions
	nameWithoutExt, extensions := helper.ExtractFileExtensions(basename)

	// Parse pattern: {database}_{YYYYMMDD}_{HHMMSS}_{hostname}
	// Format backup standard: 3 parts terakhir adalah date, time, hostname
	parts := strings.Split(nameWithoutExt, "_")
	if len(parts) < 4 {
		return "", fmt.Errorf("format filename tidak valid (minimal 4 parts): %s", nameWithoutExt)
	}

	// Ambil 3 parts terakhir: date, time, hostname
	hostname := parts[len(parts)-1]
	timeStr := parts[len(parts)-2]
	dateStr := parts[len(parts)-3]

	// Sisanya adalah database name (token {database} pada backup filename)
	dbNameParts := parts[:len(parts)-3]
	dbName := strings.Join(dbNameParts, "_")

	s.Log.Debugf("Parsed - DB: %s, Date: %s, Time: %s, Host: %s", dbName, dateStr, timeStr, hostname)

	// Build companion database name.
	// Jika backup dibuat dengan excludeData, generator akan menambahkan suffix "_nodata" ke token {database}.
	// Untuk companion, suffix _dmart harus berada sebelum _nodata.
	companionDBName := dbName + consts.SuffixDmart
	lowerDBName := strings.ToLower(dbName)
	if strings.HasSuffix(lowerDBName, "_nodata") {
		base := dbName[:len(dbName)-len("_nodata")]
		companionDBName = base + consts.SuffixDmart + "_nodata"
	}

	// List all files in directory dengan pattern yang sesuai
	// Cari: {companionDBName}_{dateStr}_*_{hostname}.{extensions}
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("gagal baca direktori: %w", err)
	}

	// Rebuild extensions untuk matching
	extStr := strings.Join(extensions, "")

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()

		// Check if matches companion pattern
		// Prefix: companionDBName_dateStr_
		prefix := companionDBName + "_" + dateStr + "_"
		if !strings.HasPrefix(filename, prefix) {
			continue
		}

		// Check if ends with hostname + extensions
		suffix := "_" + hostname + extStr
		if !strings.HasSuffix(filename, suffix) {
			continue
		}

		// Found match!
		fullPath := filepath.Join(dir, filename)
		s.Log.Debugf("Matched companion file: %s", filename)
		return fullPath, nil
	}

	return "", fmt.Errorf("tidak ada file companion ditemukan dengan pattern: %s_%s_*_%s%s",
		companionDBName, dateStr, hostname, extStr)
}

// selectCompanionFileInteractive meminta user memilih file companion database
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

	return filepath.Join(dir, files[choice-1]), nil
}
