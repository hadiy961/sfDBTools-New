package profile

import (
	"fmt"
	"os"
	"path/filepath"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper/profileutil"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

func SelectExistingDBConfig(configDir, purpose string) (types.ProfileInfo, error) {
	ui.PrintSubHeader(purpose)

	ProfileInfo := types.ProfileInfo{}

	files, err := fsops.ReadDirFiles(configDir)
	if err != nil {
		return ProfileInfo, fmt.Errorf("gagal membaca direktori konfigurasi '%s': %w", configDir, err)
	}

	filtered := make([]string, 0, len(files))
	for _, f := range files {
		if validation.ProfileExt(f) == f {
			filtered = append(filtered, f)
		}
	}

	if len(filtered) == 0 {
		ui.PrintWarning("Tidak ditemukan file konfigurasi di direktori: " + configDir)
		ui.PrintInfo("Silakan buat file konfigurasi baru terlebih dahulu dengan perintah 'profile create'.")
		return ProfileInfo, fmt.Errorf("tidak ada file konfigurasi untuk dipilih")
	}

	options := make([]string, 0, len(filtered))
	options = append(options, filtered...)

	idx, err := input.ShowMenu("Pilih file konfigurasi :", options)
	if err != nil {
		return ProfileInfo, validation.HandleInputError(err)
	}

	selected := options[idx-1]
	name := profileutil.TrimProfileSuffix(selected)

	filePath := filepath.Join(configDir, selected)
	ProfileInfo.Path = filePath
	ProfileInfo.Name = name

	info, err := LoadAndParseProfile(filePath, ProfileInfo.EncryptionKey)
	if err != nil {
		return ProfileInfo, err
	}
	if info != nil {
		ProfileInfo.DBInfo = info.DBInfo
		ProfileInfo.SSHTunnel = info.SSHTunnel
		ProfileInfo.EncryptionSource = info.EncryptionSource
	}

	var fileSizeStr string
	lastModTime := ProfileInfo.LastModified
	if fi, err := os.Stat(filePath); err == nil {
		fileSizeStr = fmt.Sprintf("%d bytes", fi.Size())
		lastModTime = fi.ModTime()
	}

	ProfileInfo.DBInfo = info.DBInfo
	ProfileInfo.SSHTunnel = info.SSHTunnel
	ProfileInfo.Size = fileSizeStr
	ProfileInfo.LastModified = lastModTime

	return ProfileInfo, nil
}
