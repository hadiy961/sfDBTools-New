// File : internal/app/profile/setup.go
// Deskripsi : Setup, configuration helpers, and path management
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 5 Januari 2026
package profile

import (
	"fmt"
	"os"
	"path/filepath"
	profilehelper "sfdbtools/internal/app/profile/helpers"
	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"strings"
)

// filePathInConfigDir membangun absolute path di dalam config dir untuk nama file konfigurasi yang diberikan.
func (s *Service) filePathInConfigDir(name string) string {
	cfgDir := s.Config.ConfigDir.DatabaseProfile
	return filepath.Join(cfgDir, shared.BuildProfileFileName(name))
}

// loadSnapshotFromPath membaca file terenkripsi, mencoba dekripsi, parse, dan mengisi OriginalProfileInfo.
func (s *Service) loadSnapshotFromPath(absPath string) error {
	info, err := profilehelper.ResolveAndLoadProfile(profilehelper.ProfileLoadOptions{
		ConfigDir:      s.Config.ConfigDir.DatabaseProfile,
		ProfilePath:    absPath,
		ProfileKey:     s.ProfileInfo.EncryptionKey,
		RequireProfile: true,
	})
	if err != nil {
		s.fillOriginalInfoFromMeta(absPath, domain.ProfileInfo{})
		return err
	}
	s.fillOriginalInfoFromMeta(absPath, *info)
	return nil
}

// fillOriginalInfoFromMeta mengisi OriginalProfileInfo dengan metadata file dan nilai koneksi yang tersedia
func (s *Service) fillOriginalInfoFromMeta(absPath string, info domain.ProfileInfo) {
	var fileSizeStr string
	var lastMod = info.LastModified
	if fi, err := os.Stat(absPath); err == nil {
		fileSizeStr = fmt.Sprintf(consts.ProfileFmtFileSizeBytes, fi.Size())
		lastMod = fi.ModTime()
	}

	s.OriginalProfileInfo = &domain.ProfileInfo{
		Path:         absPath,
		Name:         profilehelper.TrimProfileSuffix(filepath.Base(absPath)),
		DBInfo:       info.DBInfo,
		SSHTunnel:    info.SSHTunnel,
		Size:         fileSizeStr,
		LastModified: lastMod,
	}
}

// formatConfigToINI mengubah struct DBConfigInfo menjadi format string INI.
func (s *Service) formatConfigToINI() string {
	// [client] adalah header standar untuk file my.cnf
	content := `[client]
host=%s
port=%d
user=%s
password=%s
`

	base := fmt.Sprintf(content, s.ProfileInfo.DBInfo.Host, s.ProfileInfo.DBInfo.Port, s.ProfileInfo.DBInfo.User, s.ProfileInfo.DBInfo.Password)

	ssh := s.ProfileInfo.SSHTunnel
	if !ssh.Enabled && strings.TrimSpace(ssh.Host) == "" {
		return base
	}

	if ssh.Port == 0 {
		ssh.Port = 22
	}

	sshContent := `
[ssh]
enabled=%t
host=%s
port=%d
user=%s
ssh_password=%s
identity_file=%s
local_port=%d
`
	return base + fmt.Sprintf(sshContent,
		ssh.Enabled,
		ssh.Host,
		ssh.Port,
		ssh.User,
		ssh.Password,
		ssh.IdentityFile,
		ssh.LocalPort,
	)
}
