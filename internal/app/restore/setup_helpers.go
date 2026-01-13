// File : internal/restore/setup_helpers.go
// Deskripsi : Helper functions untuk setup restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 14 Januari 2026

package restore

import (
	"path/filepath"
	backupfile "sfdbtools/internal/app/backup/helpers/file"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/validation"
	"strings"
)

// extractSecondaryInstances mengekstrak instance dari list database secondary
func extractSecondaryInstances(databases []string, secondaryPrefix string) []string {
	instances := make([]string, 0)
	for _, db := range databases {
		if !strings.HasPrefix(db, secondaryPrefix) {
			continue
		}
		inst := strings.TrimPrefix(db, secondaryPrefix)
		if inst == "" {
			continue
		}
		instances = append(instances, inst)
	}
	return validation.UniqueStrings(instances)
}

// extractDefaultClientCodeFromFile mengekstrak default client code dari filename
func extractDefaultClientCodeFromFile(filePath string) string {
	inferredDB := backupfile.ExtractDatabaseNameFromFile(filepath.Base(filePath))
	inferredLower := strings.ToLower(strings.TrimSpace(inferredDB))

	if strings.HasPrefix(inferredLower, consts.PrimaryPrefixNBC) {
		return strings.TrimPrefix(inferredLower, consts.PrimaryPrefixNBC)
	} else if strings.HasPrefix(inferredLower, consts.PrimaryPrefixBiznet) {
		return strings.TrimPrefix(inferredLower, consts.PrimaryPrefixBiznet)
	}

	return ""
}
