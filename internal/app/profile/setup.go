// File : internal/app/profile/setup.go
// Deskripsi : Setup, configuration helpers, and path management
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 14 Januari 2026
package profile

import (
	"fmt"
	"path/filepath"
	"sfdbtools/internal/app/profile/helpers/snapshot"
	"sfdbtools/internal/app/profile/merger"
	"strings"
)

// filePathInConfigDir membangun absolute path di dalam config dir untuk nama file konfigurasi yang diberikan.
func (s *executorOps) filePathInConfigDir(name string) string {
	cfgDir := s.Config.ConfigDir.DatabaseProfile
	return filepath.Join(cfgDir, merger.BuildProfileFileName(name))
}

// loadSnapshotFromPath membaca file terenkripsi, mencoba dekripsi, parse, dan mengisi OriginalProfileInfo.
func (s *executorOps) loadSnapshotFromPath(absPath string) error {
	key := ""
	if s.State.ProfileInfo != nil {
		key = s.State.ProfileInfo.EncryptionKey
	}

	snap, err := snapshot.LoadProfileSnapshotFromPath(snapshot.SnapshotLoadOptions{
		ConfigDir:      s.Config.ConfigDir.DatabaseProfile,
		ProfilePath:    absPath,
		ProfileKey:     key,
		RequireProfile: true,
	})
	s.State.OriginalProfileInfo = snap
	if snap != nil {
		s.State.OriginalProfileName = snap.Name
	}
	return err
}

// formatConfigToINI mengubah struct DBConfigInfo menjadi format string INI.
func (s *executorOps) formatConfigToINI() string {
	// [client] adalah header standar untuk file my.cnf
	content := `[client]
host=%s
port=%d
user=%s
password=%s
`

	base := fmt.Sprintf(content, s.State.ProfileInfo.DBInfo.Host, s.State.ProfileInfo.DBInfo.Port, s.State.ProfileInfo.DBInfo.User, s.State.ProfileInfo.DBInfo.Password)

	ssh := s.State.ProfileInfo.SSHTunnel
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
