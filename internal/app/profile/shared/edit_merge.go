// File : internal/app/profile/shared/edit_merge.go
// Deskripsi : (DEPRECATED) Facade merge profile (snapshot/override)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 14 Januari 2026

package shared

import (
	"sfdbtools/internal/app/profile/merger"
	"sfdbtools/internal/domain"
)

func BuildProfileFileName(name string) string {
	return merger.BuildProfileFileName(name)
}

func CloneAsOriginalProfileInfo(info *domain.ProfileInfo) *domain.ProfileInfo {
	return merger.CloneAsOriginalProfileInfo(info)
}

func ApplySnapshotAsBaseline(dst *domain.ProfileInfo, snapshot *domain.ProfileInfo) {
	merger.ApplySnapshotAsBaseline(dst, snapshot)
}

func ApplyDBOverrides(dst *domain.ProfileInfo, override domain.DBInfo) {
	merger.ApplyDBOverrides(dst, override)
}

func ApplySSHOverrides(dst *domain.ProfileInfo, override domain.SSHTunnelConfig) {
	merger.ApplySSHOverrides(dst, override)
}

func HasAnyDBOverride(override domain.DBInfo) bool {
	return merger.HasAnyDBOverride(override)
}

func HasAnySSHOverride(override domain.SSHTunnelConfig) bool {
	return merger.HasAnySSHOverride(override)
}
