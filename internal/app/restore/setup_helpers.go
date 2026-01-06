// File : internal/restore/setup_helpers.go
// Deskripsi : Helper functions untuk setup restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 30 Desember 2025

package restore

import (
	"path/filepath"
	"sfdbtools/pkg/consts"
	"sfdbtools/pkg/helper"
	"strings"
)

// hasAnySuffix memeriksa apakah path memiliki salah satu suffix yang diberikan
func hasAnySuffix(path string, suffixes []string) bool {
	lower := strings.ToLower(strings.TrimSpace(path))
	for _, s := range suffixes {
		if strings.HasSuffix(lower, strings.ToLower(s)) {
			return true
		}
	}
	return false
}

// uniqueStrings mengembalikan list string unik
func uniqueStrings(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, it := range items {
		it = strings.TrimSpace(it)
		if it == "" {
			continue
		}
		if _, ok := seen[it]; ok {
			continue
		}
		seen[it] = struct{}{}
		out = append(out, it)
	}
	return out
}

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
	return uniqueStrings(instances)
}

// secondaryDBName membentuk nama database secondary dari primary dan instance
func secondaryDBName(primaryDB string, instance string) string {
	inst := strings.TrimSpace(instance)
	return primaryDB + "_secondary_" + inst
}

// extractClientCodeFromDB mengekstrak client code dari nama database
func extractClientCodeFromDB(dbName, prefix string) string {
	defaultClientCode := strings.ToLower(strings.TrimSpace(dbName))
	if strings.HasPrefix(defaultClientCode, consts.PrimaryPrefixNBC) {
		return strings.TrimPrefix(defaultClientCode, consts.PrimaryPrefixNBC)
	} else if strings.HasPrefix(defaultClientCode, consts.PrimaryPrefixBiznet) {
		return strings.TrimPrefix(defaultClientCode, consts.PrimaryPrefixBiznet)
	}
	return defaultClientCode
}

// extractDefaultClientCodeFromFile mengekstrak default client code dari filename
func extractDefaultClientCodeFromFile(filePath string) string {
	inferredDB := helper.ExtractDatabaseNameFromFile(filepath.Base(filePath))
	inferredLower := strings.ToLower(strings.TrimSpace(inferredDB))

	if strings.HasPrefix(inferredLower, consts.PrimaryPrefixNBC) {
		return strings.TrimPrefix(inferredLower, consts.PrimaryPrefixNBC)
	} else if strings.HasPrefix(inferredLower, consts.PrimaryPrefixBiznet) {
		return strings.TrimPrefix(inferredLower, consts.PrimaryPrefixBiznet)
	}

	return ""
}
