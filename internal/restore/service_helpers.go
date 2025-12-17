// File : internal/restore/service_helpers.go
// Deskripsi : Helper methods untuk Service
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package restore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"strings"
	"time"
)

// BackupTargetDatabase melakukan backup database target menggunakan backup service
func (s *Service) BackupTargetDatabase(ctx context.Context, dbName string, backupOpts *types.RestoreBackupOptions) (string, error) {
	// Determine output directory
	outputDir := ""
	if backupOpts != nil && backupOpts.OutputDir != "" {
		outputDir = backupOpts.OutputDir
	} else {
		outputDir = s.Config.Backup.Output.BaseDirectory
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat direktori output: %w", err)
	}

	// Generate filename
	timestamp := time.Now().Format("20060102_150405")
	hostname := s.Profile.DBInfo.HostName
	if hostname == "" {
		hostname = s.Profile.DBInfo.Host
	}

	filename := fmt.Sprintf("%s_%s_%s_pre_restore", dbName, timestamp, hostname)
	fullFilename := filename + ".sql"

	if backupOpts.Compression.Enabled {
		ext := compress.GetFileExtension(compress.CompressionType(backupOpts.Compression.Type))
		fullFilename += ext
	}
	if backupOpts.Encryption.Enabled {
		fullFilename += ".enc"
	}

	outputPath := filepath.Join(outputDir, fullFilename)

	// Prepare backup options
	backupOptions := &types_backup.BackupDBOptions{
		Profile:   *s.Profile,
		OutputDir: outputDir,
		Mode:      "single",
		File: types_backup.BackupFileInfo{
			Filename: filename,
		},
		Compression: types_backup.CompressionOptions{
			Enabled: backupOpts.Compression.Enabled,
			Type:    backupOpts.Compression.Type,
			Level:   backupOpts.Compression.Level,
		},
		Encryption: types_backup.EncryptionOptions{
			Enabled: backupOpts.Encryption.Enabled,
			Key:     backupOpts.Encryption.Key,
		},
		Filter: types.FilterOptions{
			IncludeDatabases: []string{dbName},
		},
	}

	// Create backup service and execute
	backupSvc := backup.NewBackupService(s.Log, s.Config, backupOptions)
	backupConfig := types_backup.BackupExecutionConfig{
		DBName:       dbName,
		DBList:       []string{dbName},
		OutputPath:   outputPath,
		BackupType:   "single",
		TotalDBFound: 1,
		IsMultiDB:    false,
	}

	_, err := backupSvc.ExecuteAndBuildBackup(ctx, backupConfig)
	if err != nil {
		return "", fmt.Errorf("gagal backup database: %w", err)
	}

	return outputPath, nil
}

// DetectOrSelectCompanionFile mendeteksi atau meminta user memilih file companion database
func (s *Service) DetectOrSelectCompanionFile() error {
	// Jika companion file sudah di-set, skip
	if s.RestorePrimaryOpts.CompanionFile != "" {
		s.Log.Infof("Menggunakan companion file yang sudah ditentukan: %s", s.RestorePrimaryOpts.CompanionFile)
		return nil
	}

	if !s.RestorePrimaryOpts.AutoDetectDmart {
		return s.selectCompanionFileInteractive()
	}

	// Auto-detect companion file
	primaryFile := s.RestorePrimaryOpts.File
	dir := filepath.Dir(primaryFile)

	s.Log.Debugf("Mencari companion file dari primary: %s", filepath.Base(primaryFile))

	// Strategi 1: Coba baca dari metadata file (.meta.json)
	companionPath, err := s.detectCompanionFromMetadata(primaryFile)
	if err == nil && companionPath != "" {
		s.RestorePrimaryOpts.CompanionFile = companionPath
		s.Log.Infof("✓ Companion file ditemukan dari metadata: %s", filepath.Base(companionPath))
		ui.PrintSuccess(fmt.Sprintf("✓ Companion file ditemukan: %s", filepath.Base(companionPath)))
		return nil
	}
	s.Log.Debugf("Gagal detect dari metadata: %v, mencoba pattern matching", err)

	// Strategi 2: Pattern matching - cari file dengan pattern yang sesuai
	companionPath, err = s.detectCompanionByPattern(primaryFile, dir)
	if err == nil && companionPath != "" {
		s.RestorePrimaryOpts.CompanionFile = companionPath
		s.Log.Infof("✓ Companion file ditemukan via pattern: %s", filepath.Base(companionPath))
		ui.PrintSuccess(fmt.Sprintf("✓ Companion file ditemukan: %s", filepath.Base(companionPath)))
		return nil
	}
	s.Log.Debugf("Gagal detect via pattern: %v", err)

	// Not found, ask user
	ui.PrintWarning("⚠️  Companion file tidak ditemukan secara otomatis")
	return s.selectCompanionFileInteractive()
}

// detectCompanionFromMetadata mencoba mendapatkan companion file dari metadata
func (s *Service) detectCompanionFromMetadata(primaryFile string) (string, error) {
	metadataPath := primaryFile + ".meta.json"

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
			// Cari yang mengandung "_dmart" di nama database
			if strings.Contains(strings.ToLower(detail.DatabaseName), "_dmart") {
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
	nameWithoutExt, extensions := s.extractExtensions(basename)

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

	// Sisanya adalah database name
	dbNameParts := parts[:len(parts)-3]
	dbName := strings.Join(dbNameParts, "_")

	s.Log.Debugf("Parsed - DB: %s, Date: %s, Time: %s, Host: %s", dbName, dateStr, timeStr, hostname)

	// Build companion database name
	companionDBName := dbName + "_dmart"

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

// extractExtensions mengekstrak ekstensi dari filename dan mengembalikan nama tanpa ekstensi + list ekstensi
func (s *Service) extractExtensions(filename string) (string, []string) {
	nameWithoutExt := filename
	extensions := []string{}

	// Remove .enc
	if strings.HasSuffix(strings.ToLower(nameWithoutExt), ".enc") {
		nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ".enc")
		extensions = append([]string{".enc"}, extensions...)
	}

	// Remove compression extension
	for _, ext := range []string{".gz", ".xz", ".zst", ".bz2"} {
		if strings.HasSuffix(strings.ToLower(nameWithoutExt), ext) {
			nameWithoutExt = nameWithoutExt[:len(nameWithoutExt)-len(ext)]
			extensions = append([]string{ext}, extensions...)
			break
		}
	}

	// Remove .sql
	if strings.HasSuffix(strings.ToLower(nameWithoutExt), ".sql") {
		nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ".sql")
		extensions = append([]string{".sql"}, extensions...)
	}

	return nameWithoutExt, extensions
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

	// Show file selector
	dir := filepath.Dir(s.RestorePrimaryOpts.File)
	files, err := helper.ListBackupFilesInDirectory(dir)
	if err != nil {
		return fmt.Errorf("gagal list files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("tidak ada file backup ditemukan di direktori: %s", dir)
	}

	choice, err := input.ShowMenu("Pilih file companion database:", files)
	if err != nil {
		return fmt.Errorf("gagal memilih file: %w", err)
	}

	s.RestorePrimaryOpts.CompanionFile = filepath.Join(dir, files[choice-1])
	s.Log.Infof("User memilih companion file: %s", files[choice-1])

	return nil
}

// validateApplicationPassword memvalidasi password aplikasi sebelum restore primary
func (s *Service) validateApplicationPassword() error {
	s.Log.Info("Meminta password aplikasi untuk validasi restore primary")

	// Prompt password
	password, err := input.PromptPassword("Masukkan password aplikasi untuk melanjutkan restore primary:")
	if err != nil {
		return fmt.Errorf("gagal membaca password: %w", err)
	}

	// Validasi password dengan ENV_PASSWORD_APP dari consts
	if password != consts.ENV_PASSWORD_APP {
		s.Log.Error("Password aplikasi tidak valid")
		return fmt.Errorf("password aplikasi tidak valid")
	}

	s.Log.Info("Password aplikasi valid, melanjutkan restore")
	ui.PrintSuccess("✓ Password aplikasi valid")

	return nil
}
