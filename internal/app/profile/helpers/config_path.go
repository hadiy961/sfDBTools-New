package helpers

import (
	"fmt"
	"path/filepath"
	"strings"

	appconfig "sfdbtools/internal/services/config"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/validation"
)

// TrimProfileSuffix menghapus suffix .cnf.enc dari nama jika ada.
func TrimProfileSuffix(name string) string {
	return strings.TrimSuffix(strings.TrimSuffix(name, consts.ExtEnc), consts.ExtCnf)
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

	cfgDir := cfg.ConfigDir.DatabaseProfile
	var absPath string
	if filepath.IsAbs(spec) {
		absPath = filepath.Clean(spec)
	} else {
		// Untuk input relatif, hanya izinkan base filename agar tidak bisa path traversal.
		// (custom lokasi file bisa pakai absolute path atau flag output-dir saat create)
		if err := validation.ValidateCustomFilenameBase(spec); err != nil {
			return "", "", err
		}
		absPath = filepath.Join(cfgDir, spec)
	}
	absPath = validation.ProfileExt(absPath)

	name := TrimProfileSuffix(filepath.Base(absPath))
	return absPath, name, nil
}
