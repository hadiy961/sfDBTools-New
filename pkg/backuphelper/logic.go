// File : pkg/backuphelper/logic.go
// Deskripsi : Pure logic helpers untuk backup operations yang independen dari service state
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-11
// Last Modified : 2025-12-11

package backuphelper

import (
	"strings"
)

// ExtractMysqldumpVersion mengambil versi mysqldump dari stderr output
// Biasanya format: "mysqldump  Ver 10.19 Distrib 10.11.6-MariaDB, for Linux (x86_64)"
func ExtractMysqldumpVersion(stderrOutput string) string {
	if stderrOutput == "" {
		return ""
	}

	lines := strings.Split(stderrOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "mysqldump") && strings.Contains(line, "Ver") {
			// Extract version from line like "mysqldump  Ver 10.19 Distrib 10.11.6-MariaDB"
			return strings.TrimSpace(line)
		}
	}

	return ""
}

// FilterCandidatesByMode memfilter database candidates berdasarkan backup mode
func FilterCandidatesByMode(dbFiltered []string, mode string) []string {
	candidates := make([]string, 0, len(dbFiltered))

	// System databases yang harus di-exclude
	systemDatabases := []string{"information_schema", "performance_schema", "mysql", "sys"}

	for _, db := range dbFiltered {
		dbLower := strings.ToLower(db)

		// Check if it's a system database (untuk semua mode)
		isSystemDB := false
		for _, sysDB := range systemDatabases {
			if dbLower == sysDB {
				isSystemDB = true
				break
			}
		}
		if isSystemDB {
			continue
		}

		switch mode {
		case "primary":
			// Primary: hanya database dengan pattern dbsf_nbc_{nama_database}
			// Exclude: yang punya _secondary, _dmart, _temp, _archive
			if strings.Contains(dbLower, "_secondary") || strings.HasSuffix(dbLower, "_dmart") ||
				strings.HasSuffix(dbLower, "_temp") || strings.HasSuffix(dbLower, "_archive") {
				continue
			}
			// Harus match pattern dbsf_nbc_ 
			if !strings.HasPrefix(dbLower, "dbsf_nbc_") {
				continue
			}
		case "secondary":
			// Secondary: hanya database dengan pattern dbsf_nbc_{nama_database}_secondary_{instance}
			// Exclude: _dmart, _temp, _archive
			if strings.HasSuffix(dbLower, "_dmart") || strings.HasSuffix(dbLower, "_temp") ||
				strings.HasSuffix(dbLower, "_archive") {
				continue
			}
			// Harus match pattern dbsf_nbc_ dan mengandung _secondary
			if !strings.HasPrefix(dbLower, "dbsf_nbc_") || !strings.Contains(dbLower, "_secondary") {
				continue
			}
		case "single":
			// Single: exclude companion databases
			if strings.HasSuffix(dbLower, "_dmart") || strings.HasSuffix(dbLower, "_temp") ||
				strings.HasSuffix(dbLower, "_archive") {
				continue
			}
		}

		candidates = append(candidates, db)
	}

	return candidates
}

// IsSingleModeVariant checks if mode is single/primary/secondary
func IsSingleModeVariant(mode string) bool {
	return mode == "single" || mode == "primary" || mode == "secondary"
}
