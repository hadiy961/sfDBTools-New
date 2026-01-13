// File : internal/app/profile/helpers/profile_snapshot.go
// Deskripsi : Helper untuk load snapshot profile + metadata file (untuk show/edit)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
)

type SnapshotLoadOptions struct {
	ConfigDir      string
	ProfilePath    string
	ProfileKey     string
	RequireProfile bool
}

func buildSnapshotFromMeta(absPath string, info *domain.ProfileInfo) *domain.ProfileInfo {
	var (
		fileSizeStr string
		lastMod     = info.LastModified
	)
	if fi, err := os.Stat(absPath); err == nil {
		fileSizeStr = fmt.Sprintf(consts.ProfileFmtFileSizeBytes, fi.Size())
		lastMod = fi.ModTime()
	}

	name := shared.TrimProfileSuffix(filepath.Base(absPath))
	return &domain.ProfileInfo{
		Path:         absPath,
		Name:         name,
		DBInfo:       info.DBInfo,
		SSHTunnel:    info.SSHTunnel,
		Size:         fileSizeStr,
		LastModified: lastMod,
		// EncryptionSource sengaja tidak diset di snapshot; ini fokus untuk display baseline.
	}
}

// LoadProfileSnapshotFromPath membaca file profile, mencoba dekripsi+parse, lalu mengembalikan snapshot
// yang selalu berisi metadata file (size/mtime) dan baseline value (jika load berhasil).
// Jika load gagal, snapshot tetap dikembalikan (dengan DBInfo kosong) bersama error.
func LoadProfileSnapshotFromPath(opts SnapshotLoadOptions) (*domain.ProfileInfo, error) {
	if strings.TrimSpace(opts.ProfilePath) == "" {
		return nil, shared.ErrProfilePathEmpty
	}

	loaded, err := ResolveAndLoadProfile(ProfileLoadOptions{
		ConfigDir:      opts.ConfigDir,
		ProfilePath:    opts.ProfilePath,
		ProfileKey:     opts.ProfileKey,
		RequireProfile: opts.RequireProfile,
	})
	if err != nil {
		return buildSnapshotFromMeta(opts.ProfilePath, &domain.ProfileInfo{}), err
	}
	if loaded == nil {
		return buildSnapshotFromMeta(opts.ProfilePath, &domain.ProfileInfo{}), fmt.Errorf("profile berhasil di-resolve tapi hasil load nil")
	}
	return buildSnapshotFromMeta(opts.ProfilePath, loaded), nil
}
