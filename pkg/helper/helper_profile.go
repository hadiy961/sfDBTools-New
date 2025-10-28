package helper

import (
	"fmt"
	"path/filepath"
	"sfDBTools/internal/appconfig"
	"sfDBTools/pkg/validation"
	"strings"
)

// ResolveConfigPath mengubah name/path menjadi absolute path di config dir dan nama normalized tanpa suffix
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
		absPath = spec
	} else {
		absPath = filepath.Join(cfgDir, spec)
	}
	absPath = validation.ProfileExt(absPath)

	// Nama normalized
	name := TrimProfileSuffix(filepath.Base(absPath))
	return absPath, name, nil
}
