package helpers

import (
	"fmt"
	"path/filepath"
	"strings"

	"sfdbtools/internal/app/profile/shared"
	appconfig "sfdbtools/internal/services/config"
	"sfdbtools/internal/shared/validation"
)

// ResolveConfigPathInDir mengubah name/path menjadi absolute path di configDir dan nama normalized tanpa suffix.
// Tidak memuat config.yaml (lebih efisien untuk dipanggil berulang).
func ResolveConfigPathInDir(configDir string, spec string) (string, string, error) {
	if strings.TrimSpace(spec) == "" {
		return "", "", fmt.Errorf("nama atau path file konfigurasi kosong")
	}

	var absPath string
	if filepath.IsAbs(spec) {
		absPath = filepath.Clean(spec)
	} else {
		if strings.TrimSpace(configDir) == "" {
			return "", "", fmt.Errorf("configDir kosong untuk resolve path relatif: %s", spec)
		}
		// Untuk input relatif, hanya izinkan base filename agar tidak bisa path traversal.
		if err := validation.ValidateCustomFilenameBase(spec); err != nil {
			return "", "", err
		}
		absPath = filepath.Join(configDir, spec)
	}
	absPath = validation.ProfileExt(absPath)

	name := shared.TrimProfileSuffix(filepath.Base(absPath))
	return absPath, name, nil
}

// ResolveConfigPath mengubah name/path menjadi absolute path di config dir dan nama normalized tanpa suffix.
func ResolveConfigPath(spec string) (string, string, error) {
	if strings.TrimSpace(spec) == "" {
		return "", "", fmt.Errorf("nama atau path file konfigurasi kosong")
	}

	cfg, err := appconfig.LoadConfigFromEnv()
	if err != nil {
		return "", "", fmt.Errorf("gagal memuat konfigurasi aplikasi: %w", err)
	}

	return ResolveConfigPathInDir(cfg.ConfigDir.DatabaseProfile, spec)
}
