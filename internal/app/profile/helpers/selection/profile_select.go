// File : internal/app/profile/helpers/selection/profile_select.go
// Deskripsi : Helper untuk pilih profile (interactive) + snapshot (untuk show/edit)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package selection

import (
	"fmt"
	"os"
	"path/filepath"

	"sfdbtools/internal/app/profile/helpers/parser"
	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/fsops"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
)

func SelectExistingDBConfig(configDir, purpose string) (domain.ProfileInfo, error) {
	print.PrintSubHeader(purpose)

	profileInfo := domain.ProfileInfo{}

	files, err := fsops.ReadDirFiles(configDir)
	if err != nil {
		return profileInfo, fmt.Errorf("gagal membaca direktori konfigurasi '%s': %w", configDir, err)
	}

	filtered := make([]string, 0, len(files))
	for _, f := range files {
		if validation.ProfileExt(f) == f {
			filtered = append(filtered, f)
		}
	}

	if len(filtered) == 0 {
		print.PrintWarn("Tidak ditemukan file konfigurasi di direktori: " + configDir)
		print.PrintInfo("Silakan buat file konfigurasi baru terlebih dahulu dengan perintah 'profile create'.")
		return profileInfo, shared.ErrNoConfigToSelect
	}

	options := filtered

	selected, _, err := prompt.SelectOne("Pilih file konfigurasi :", options, 0)
	if err != nil {
		return profileInfo, validation.HandleInputError(err)
	}
	name := shared.TrimProfileSuffix(selected)

	filePath := filepath.Join(configDir, selected)
	profileInfo.Path = filePath
	profileInfo.Name = name

	info, err := parser.LoadAndParseProfile(filePath, profileInfo.EncryptionKey)
	if err != nil {
		return profileInfo, err
	}
	profileInfo.DBInfo = info.DBInfo
	profileInfo.SSHTunnel = info.SSHTunnel
	profileInfo.EncryptionSource = info.EncryptionSource

	var fileSizeStr string
	lastModTime := profileInfo.LastModified
	if fi, err := os.Stat(filePath); err == nil {
		fileSizeStr = fmt.Sprintf("%d bytes", fi.Size())
		lastModTime = fi.ModTime()
	}

	profileInfo.Size = fileSizeStr
	profileInfo.LastModified = lastModTime

	return profileInfo, nil
}
