// File : internal/backup/helpers/logic.go
// Deskripsi : Pure logic helpers untuk backup operations yang independen dari service state
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-11
// Last Modified : 2025-12-11

package helpers

import (
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
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

	for _, db := range dbFiltered {
		dbLower := strings.ToLower(db)

		// Check if it's a system database (untuk semua mode)
		if _, isSystemDB := types.SystemDatabases[dbLower]; isSystemDB {
			continue
		}

		switch mode {
		case consts.ModePrimary:
			// Primary: hanya database dengan pattern dbsf_nbc_{nama_database}
			// Exclude: yang punya _secondary, _dmart, _temp, _archive
			if strings.Contains(dbLower, consts.SecondarySuffix) || strings.HasSuffix(dbLower, consts.SuffixDmart) ||
				strings.HasSuffix(dbLower, consts.SuffixTemp) || strings.HasSuffix(dbLower, consts.SuffixArchive) {
				continue
			}
			// Harus match pattern dbsf_nbc_
			if !strings.HasPrefix(dbLower, consts.PrimaryPrefixNBC) {
				continue
			}
		case consts.ModeSecondary:
			// Secondary: hanya database dengan pattern dbsf_nbc_{nama_database}_secondary_{instance}
			// Exclude: _dmart, _temp, _archive
			if strings.HasSuffix(dbLower, consts.SuffixDmart) || strings.HasSuffix(dbLower, consts.SuffixTemp) ||
				strings.HasSuffix(dbLower, consts.SuffixArchive) {
				continue
			}
			// Harus match pattern dbsf_nbc_ dan mengandung _secondary
			if !strings.HasPrefix(dbLower, consts.PrimaryPrefixNBC) || !strings.Contains(dbLower, consts.SecondarySuffix) {
				continue
			}
		case consts.ModeSingle:
			// Single: exclude companion databases
			if strings.HasSuffix(dbLower, consts.SuffixDmart) || strings.HasSuffix(dbLower, consts.SuffixTemp) ||
				strings.HasSuffix(dbLower, consts.SuffixArchive) {
				continue
			}
		}

		candidates = append(candidates, db)
	}

	return candidates
}

// FilterCandidatesByClientCode memfilter database berdasarkan client code untuk mode primary
// Pattern: dbsf_nbc_{client_code}
func FilterCandidatesByClientCode(databases []string, clientCode string) []string {
	if clientCode == "" {
		return databases
	}

	filtered := make([]string, 0)
	targetPattern := consts.PrimaryPrefixNBC + strings.ToLower(clientCode)

	for _, db := range databases {
		dbLower := strings.ToLower(db)
		// Database harus match dbsf_nbc_{client_code} atau dbsf_nbc_{client_code}_*
		if dbLower == targetPattern || strings.HasPrefix(dbLower, targetPattern+"_") {
			// Tapi exclude yang punya _secondary, _dmart, _temp, _archive
			if !strings.Contains(dbLower, consts.SecondarySuffix) &&
				!strings.HasSuffix(dbLower, consts.SuffixDmart) &&
				!strings.HasSuffix(dbLower, consts.SuffixTemp) &&
				!strings.HasSuffix(dbLower, consts.SuffixArchive) {
				filtered = append(filtered, db)
			}
		}
	}

	return filtered
}

// FilterSecondaryByClientCodeAndInstance memfilter database secondary berdasarkan client code dan instance
// Pattern: dbsf_nbc_{client_code}_secondary_{instance}
func FilterSecondaryByClientCodeAndInstance(databases []string, clientCode, instance string) []string {
	filtered := make([]string, 0)

	// Jika hanya instance (tanpa client code)
	if clientCode == "" && instance != "" {
		targetSuffix := consts.SecondarySuffix + "_" + strings.ToLower(instance)
		for _, db := range databases {
			dbLower := strings.ToLower(db)
			// Database harus mengandung _secondary_{instance}
			if strings.Contains(dbLower, targetSuffix) {
				// Exclude companion databases
				if !strings.HasSuffix(dbLower, consts.SuffixDmart) &&
					!strings.HasSuffix(dbLower, consts.SuffixTemp) &&
					!strings.HasSuffix(dbLower, consts.SuffixArchive) {
					filtered = append(filtered, db)
				}
			}
		}
		return filtered
	}

	// Jika hanya client code
	if clientCode != "" && instance == "" {
		targetPattern := consts.PrimaryPrefixNBC + strings.ToLower(clientCode) + consts.SecondarySuffix
		for _, db := range databases {
			dbLower := strings.ToLower(db)
			if strings.Contains(dbLower, targetPattern) {
				// Exclude companion databases
				if !strings.HasSuffix(dbLower, consts.SuffixDmart) &&
					!strings.HasSuffix(dbLower, consts.SuffixTemp) &&
					!strings.HasSuffix(dbLower, consts.SuffixArchive) {
					filtered = append(filtered, db)
				}
			}
		}
		return filtered
	}

	// Jika client code dan instance
	if clientCode != "" && instance != "" {
		targetPattern := consts.PrimaryPrefixNBC + strings.ToLower(clientCode) + consts.SecondarySuffix + "_" + strings.ToLower(instance)
		for _, db := range databases {
			dbLower := strings.ToLower(db)
			// Exact match atau dengan suffix lain (tapi bukan companion)
			if dbLower == targetPattern || strings.HasPrefix(dbLower, targetPattern+"_") {
				// Exclude companion databases
				if !strings.HasSuffix(dbLower, consts.SuffixDmart) &&
					!strings.HasSuffix(dbLower, consts.SuffixTemp) &&
					!strings.HasSuffix(dbLower, consts.SuffixArchive) {
					filtered = append(filtered, db)
				}
			}
		}
		return filtered
	}

	return databases
}

// IsSingleModeVariant checks if mode is single/primary/secondary
func IsSingleModeVariant(mode string) bool {
	return mode == consts.ModeSingle || mode == consts.ModePrimary || mode == consts.ModeSecondary
}
