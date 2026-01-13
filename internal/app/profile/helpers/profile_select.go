// File : internal/app/profile/helpers/profile_select.go
// Deskripsi : Helper untuk pilih profile (interactive) + snapshot (untuk show/edit)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package helpers

import (
	"fmt"
	"os"
	"path/filepath"

	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
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

	info, err := LoadAndParseProfile(filePath, profileInfo.EncryptionKey)
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

// SelectExistingDBConfigWithSnapshot memilih profile secara interaktif lalu mengembalikan:
// - info (hasil load+parse profile)
// - originalName (nama profile yang dipilih)
// - snapshot (salinan info untuk baseline/original display)
func SelectExistingDBConfigWithSnapshot(configDir string, prompt string) (info *domain.ProfileInfo, originalName string, snapshot *domain.ProfileInfo, err error) {
	loaded, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ConfigDir:         configDir,
		ProfilePath:       "",
		AllowInteractive:  true,
		InteractivePrompt: prompt,
		RequireProfile:    true,
	})
	if err != nil {
		return nil, "", nil, err
	}
	if loaded == nil {
		return nil, "", nil, fmt.Errorf("%sprofile tidak tersedia (hasil load nil)", consts.ProfileMsgNonInteractivePrefix)
	}

	originalName = loaded.Name
	snapshot = shared.CloneAsOriginalProfileInfo(loaded)
	return loaded, originalName, snapshot, nil
}
