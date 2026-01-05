package selection

import (
	"strings"

	"sfDBTools/internal/domain"
	"sfDBTools/pkg/consts"
)

// FilterCandidatesByMode memfilter database candidates berdasarkan backup mode.
func FilterCandidatesByMode(dbFiltered []string, mode string) []string {
	candidates := make([]string, 0, len(dbFiltered))

	for _, db := range dbFiltered {
		dbLower := strings.ToLower(db)

		if _, isSystemDB := domain.SystemDatabases[dbLower]; isSystemDB {
			continue
		}

		switch mode {
		case consts.ModePrimary:
			if strings.Contains(dbLower, consts.SecondarySuffix) || strings.HasSuffix(dbLower, consts.SuffixDmart) ||
				strings.HasSuffix(dbLower, consts.SuffixTemp) || strings.HasSuffix(dbLower, consts.SuffixArchive) {
				continue
			}
			if !strings.HasPrefix(dbLower, consts.PrimaryPrefixNBC) {
				continue
			}
		case consts.ModeSecondary:
			if strings.HasSuffix(dbLower, consts.SuffixDmart) || strings.HasSuffix(dbLower, consts.SuffixTemp) ||
				strings.HasSuffix(dbLower, consts.SuffixArchive) {
				continue
			}
			if !strings.HasPrefix(dbLower, consts.PrimaryPrefixNBC) || !strings.Contains(dbLower, consts.SecondarySuffix) {
				continue
			}
		case consts.ModeSingle:
			if strings.HasSuffix(dbLower, consts.SuffixDmart) || strings.HasSuffix(dbLower, consts.SuffixTemp) ||
				strings.HasSuffix(dbLower, consts.SuffixArchive) {
				continue
			}
		}

		candidates = append(candidates, db)
	}

	return candidates
}

// FilterCandidatesByClientCode memfilter database berdasarkan client code untuk mode primary.
// Pattern: dbsf_nbc_{client_code}
func FilterCandidatesByClientCode(databases []string, clientCode string) []string {
	if clientCode == "" {
		return databases
	}

	filtered := make([]string, 0, len(databases))
	targetPattern := consts.PrimaryPrefixNBC + strings.ToLower(clientCode)

	for _, db := range databases {
		dbLower := strings.ToLower(db)
		if dbLower == targetPattern || strings.HasPrefix(dbLower, targetPattern+"_") {
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

// FilterSecondaryByClientCodeAndInstance memfilter database secondary berdasarkan client code dan instance.
// Pattern: dbsf_nbc_{client_code}_secondary_{instance}
func FilterSecondaryByClientCodeAndInstance(databases []string, clientCode, instance string) []string {
	filtered := make([]string, 0)

	if clientCode == "" && instance != "" {
		targetSuffix := consts.SecondarySuffix + "_" + strings.ToLower(instance)
		for _, db := range databases {
			dbLower := strings.ToLower(db)
			if strings.Contains(dbLower, targetSuffix) {
				if !strings.HasSuffix(dbLower, consts.SuffixDmart) &&
					!strings.HasSuffix(dbLower, consts.SuffixTemp) &&
					!strings.HasSuffix(dbLower, consts.SuffixArchive) {
					filtered = append(filtered, db)
				}
			}
		}
		return filtered
	}

	if clientCode != "" && instance == "" {
		targetPattern := consts.PrimaryPrefixNBC + strings.ToLower(clientCode) + consts.SecondarySuffix
		for _, db := range databases {
			dbLower := strings.ToLower(db)
			if strings.Contains(dbLower, targetPattern) {
				if !strings.HasSuffix(dbLower, consts.SuffixDmart) &&
					!strings.HasSuffix(dbLower, consts.SuffixTemp) &&
					!strings.HasSuffix(dbLower, consts.SuffixArchive) {
					filtered = append(filtered, db)
				}
			}
		}
		return filtered
	}

	if clientCode != "" && instance != "" {
		targetPattern := consts.PrimaryPrefixNBC + strings.ToLower(clientCode) + consts.SecondarySuffix + "_" + strings.ToLower(instance)
		for _, db := range databases {
			dbLower := strings.ToLower(db)
			if dbLower == targetPattern || strings.HasPrefix(dbLower, targetPattern+"_") {
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
