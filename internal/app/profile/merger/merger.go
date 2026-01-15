// File : internal/app/profile/merger/merger.go
// Deskripsi : Helper untuk merge snapshot dan override saat edit profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 14 Januari 2026

package merger

import (
	"strings"

	"sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/validation"
)

func BuildProfileFileName(name string) string {
	return validation.ProfileExt(connection.TrimProfileSuffix(name))
}

func CloneAsOriginalProfileInfo(info *domain.ProfileInfo) *domain.ProfileInfo {
	if info == nil {
		return nil
	}
	return &domain.ProfileInfo{
		Path:         info.Path,
		Name:         info.Name,
		DBInfo:       info.DBInfo,
		SSHTunnel:    info.SSHTunnel,
		Size:         info.Size,
		LastModified: info.LastModified,
	}
}

func ApplySnapshotAsBaseline(dst *domain.ProfileInfo, snapshot *domain.ProfileInfo) {
	if dst == nil || snapshot == nil {
		return
	}
	dst.DBInfo = snapshot.DBInfo
	dst.SSHTunnel = snapshot.SSHTunnel
	if strings.TrimSpace(dst.Name) == "" {
		dst.Name = snapshot.Name
	}
	if strings.TrimSpace(dst.Path) == "" {
		dst.Path = snapshot.Path
	}
}

func ApplyDBOverrides(dst *domain.ProfileInfo, override domain.DBInfo) {
	if dst == nil {
		return
	}
	if strings.TrimSpace(override.Host) != "" {
		dst.DBInfo.Host = override.Host
	}
	if override.Port != 0 {
		dst.DBInfo.Port = override.Port
	}
	if strings.TrimSpace(override.User) != "" {
		dst.DBInfo.User = override.User
	}
	if strings.TrimSpace(override.Password) != "" {
		dst.DBInfo.Password = override.Password
	}
}

func ApplySSHOverrides(dst *domain.ProfileInfo, override domain.SSHTunnelConfig) {
	if dst == nil {
		return
	}
	// Enabled override hanya jika user mengaktifkan (avoid overwrite baseline dengan default false)
	if override.Enabled {
		dst.SSHTunnel.Enabled = true
	}
	if strings.TrimSpace(override.Host) != "" {
		dst.SSHTunnel.Host = override.Host
	}
	if override.Port != 0 {
		dst.SSHTunnel.Port = override.Port
	}
	if strings.TrimSpace(override.User) != "" {
		dst.SSHTunnel.User = override.User
	}
	if strings.TrimSpace(override.Password) != "" {
		dst.SSHTunnel.Password = override.Password
	}
	if strings.TrimSpace(override.IdentityFile) != "" {
		dst.SSHTunnel.IdentityFile = override.IdentityFile
	}
	if override.LocalPort != 0 {
		dst.SSHTunnel.LocalPort = override.LocalPort
	}
}

func HasAnyDBOverride(override domain.DBInfo) bool {
	return strings.TrimSpace(override.Host) != "" || override.Port != 0 ||
		strings.TrimSpace(override.User) != "" || strings.TrimSpace(override.Password) != ""
}

func HasAnySSHOverride(override domain.SSHTunnelConfig) bool {
	return override.Enabled || strings.TrimSpace(override.Host) != "" || override.Port != 0 ||
		strings.TrimSpace(override.User) != "" || strings.TrimSpace(override.Password) != "" ||
		strings.TrimSpace(override.IdentityFile) != "" || override.LocalPort != 0
}
