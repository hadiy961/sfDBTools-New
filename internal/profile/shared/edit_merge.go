// File : internal/profile/shared/edit_merge.go
// Deskripsi : Helper shared untuk merge snapshot dan override saat edit profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 4 Januari 2026

package shared

import (
	"strings"

	"sfDBTools/internal/types"
)

func ApplySnapshotAsBaseline(dst *types.ProfileInfo, snapshot *types.ProfileInfo) {
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

func ApplyDBOverrides(dst *types.ProfileInfo, override types.DBInfo) {
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

func ApplySSHOverrides(dst *types.ProfileInfo, override types.SSHTunnelConfig) {
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

func HasAnyDBOverride(override types.DBInfo) bool {
	return strings.TrimSpace(override.Host) != "" || override.Port != 0 ||
		strings.TrimSpace(override.User) != "" || strings.TrimSpace(override.Password) != ""
}

func HasAnySSHOverride(override types.SSHTunnelConfig) bool {
	return override.Enabled || strings.TrimSpace(override.Host) != "" || override.Port != 0 ||
		strings.TrimSpace(override.User) != "" || strings.TrimSpace(override.Password) != "" ||
		strings.TrimSpace(override.IdentityFile) != "" || override.LocalPort != 0
}
