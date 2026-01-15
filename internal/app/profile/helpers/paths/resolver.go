// File : internal/app/profile/helpers/paths/resolver.go
// Deskripsi : PathResolver untuk sentralisasi logic resolve path profile
// Author : Hadiyatna Muflihun
// Tanggal : 15 Januari 2026
// Last Modified : 15 Januari 2026

package paths

import (
	"fmt"
	"strings"
)

// PathResolver menyediakan method untuk resolve path profile dengan atau tanpa configDir.
// Menghilangkan duplikasi pola if-else untuk memilih ResolveConfigPath vs ResolveConfigPathInDir.
type PathResolver struct {
	ConfigDir string
}

// NewPathResolver membuat PathResolver baru dengan configDir optional.
func NewPathResolver(configDir string) *PathResolver {
	return &PathResolver{ConfigDir: configDir}
}

// Resolve mengubah spec (name/path) menjadi absolute path dan nama normalized.
// Otomatis memilih ResolveConfigPathInDir atau ResolveConfigPath berdasarkan ConfigDir.
func (r *PathResolver) Resolve(spec string) (absPath string, name string, err error) {
	if strings.TrimSpace(r.ConfigDir) != "" {
		absPath, name, err = ResolveConfigPathInDir(r.ConfigDir, spec)
	} else {
		absPath, name, err = ResolveConfigPath(spec)
	}
	if err != nil {
		return "", "", fmt.Errorf("gagal memproses path konfigurasi: %w", err)
	}
	return absPath, name, nil
}
