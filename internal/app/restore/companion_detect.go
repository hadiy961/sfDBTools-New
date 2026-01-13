// File : internal/restore/companion_detect.go
// Deskripsi : Helper deteksi companion (_dmart) file
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 14 Januari 2026

package restore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	backupfile "sfdbtools/internal/app/backup/helpers/file"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/naming"
	"strings"
)

// detectCompanionAuto mencoba menemukan companion file menggunakan strategi berurutan.
func (s *Service) detectCompanionAuto(primaryFile string) (string, error) {
	dir := filepath.Dir(primaryFile)

	s.Log.Infof(
		"Auto-detect companion (_dmart) rule: 1) baca metadata '%s', 2) jika gagal, pattern standar, 3) fallback sibling filename",
		filepath.Base(primaryFile)+consts.ExtMetaJSON,
	)
	s.Log.Info("Pattern standar: {database}_{YYYYMMDD}_{HHMMSS}_{hostname}.sql[.gz][.enc] (dan companion {database}_dmart...) ")
	s.Log.Debugf("Mencari companion file dari primary: %s", filepath.Base(primaryFile))

	companionPath, err := s.detectCompanionFromMetadata(primaryFile)
	if err == nil && companionPath != "" {
		return companionPath, nil
	}
	s.Log.Debugf("Gagal detect dari metadata: %v, mencoba pattern matching", err)

	companionPath, err = s.detectCompanionByPattern(primaryFile, dir)
	if err == nil && companionPath != "" {
		return companionPath, nil
	}
	s.Log.Debugf("Gagal detect via pattern: %v", err)

	companionPath, err = s.detectCompanionBySiblingFilename(primaryFile, dir)
	if err == nil && companionPath != "" {
		return companionPath, nil
	}
	s.Log.Debugf("Gagal detect via sibling filename: %v", err)

	if err == nil {
		err = fmt.Errorf("companion file tidak ditemukan")
	}
	return "", err
}

// detectCompanionBySiblingFilename mencoba menemukan companion file dengan pola sederhana:
//
//	primary:   dbsf_nbc_xxx[_nodata].sql(.gz/.zst/.enc)
//	companion: dbsf_nbc_xxx_dmart[_nodata].sql(.gz/.zst/.enc)
//
// Ini berguna untuk dump yang tidak memakai format timestamp/hostname dan tidak punya .meta.json.
func (s *Service) detectCompanionBySiblingFilename(primaryFile string, dir string) (string, error) {
	basename := filepath.Base(primaryFile)
	nameWithoutExt, extensions := backupfile.ExtractFileExtensions(basename)
	if nameWithoutExt == "" {
		return "", fmt.Errorf("gagal parse filename: %s", basename)
	}

	companionDBName := naming.BuildCompanionDBName(nameWithoutExt)

	extStr := strings.Join(extensions, "")
	candidate := filepath.Join(dir, companionDBName+extStr)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	return "", fmt.Errorf("companion sibling tidak ditemukan: %s", filepath.Base(candidate))
}

// detectCompanionFromMetadata mencoba mendapatkan companion file dari metadata.
func (s *Service) detectCompanionFromMetadata(primaryFile string) (string, error) {
	metadataPath := primaryFile + consts.ExtMetaJSON

	if _, err := os.Stat(metadataPath); err != nil {
		return "", fmt.Errorf("metadata file tidak ditemukan: %w", err)
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return "", fmt.Errorf("gagal baca metadata: %w", err)
	}

	var meta types_backup.BackupMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return "", fmt.Errorf("gagal parse metadata: %w", err)
	}

	for _, detail := range meta.DatabaseDetails {
		if strings.Contains(strings.ToLower(detail.DatabaseName), consts.SuffixDmart) {
			if _, err := os.Stat(detail.BackupFile); err == nil {
				s.Log.Debugf("Found companion in metadata: %s", detail.DatabaseName)
				return detail.BackupFile, nil
			}
			s.Log.Warnf("Companion file di metadata tidak ada di disk: %s", detail.BackupFile)
		}
	}

	return "", fmt.Errorf("tidak ada companion database ditemukan di metadata")
}

// detectCompanionByPattern mencoba menemukan companion file menggunakan pattern matching.
// Format backup file: {database}_{YYYYMMDD}_{HHMMSS}_{hostname}.sql.gz[.enc]
func (s *Service) detectCompanionByPattern(primaryFile string, dir string) (string, error) {
	basename := filepath.Base(primaryFile)

	nameWithoutExt, extensions := backupfile.ExtractFileExtensions(basename)

	parts := strings.Split(nameWithoutExt, "_")
	if len(parts) < 4 {
		return "", fmt.Errorf("format filename tidak valid (minimal 4 parts): %s", nameWithoutExt)
	}

	hostname := parts[len(parts)-1]
	dateStr := parts[len(parts)-3]
	dbName := strings.Join(parts[:len(parts)-3], "_")

	s.Log.Debugf("Parsed - DB: %s, Date: %s, Host: %s", dbName, dateStr, hostname)

	companionDBName := naming.BuildCompanionDBName(dbName)

	files, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("gagal baca direktori: %w", err)
	}

	extStr := strings.Join(extensions, "")
	prefix := companionDBName + "_" + dateStr + "_"
	suffix := "_" + hostname + extStr

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filename := file.Name()
		if !strings.HasPrefix(filename, prefix) {
			continue
		}
		if !strings.HasSuffix(filename, suffix) {
			continue
		}
		fullPath := filepath.Join(dir, filename)
		s.Log.Debugf("Matched companion file: %s", filename)
		return fullPath, nil
	}

	return "", fmt.Errorf(
		"tidak ada file companion ditemukan dengan pattern: %s_%s_*_%s%s",
		companionDBName,
		dateStr,
		hostname,
		extStr,
	)
}
