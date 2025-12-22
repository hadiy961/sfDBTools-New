package backup

import (
	"fmt"
	"sfDBTools/internal/backup/helpers"
	"sfDBTools/pkg/consts"
)

func (s *Service) filterCandidatesByModeAndOptions(mode string, candidates []string) ([]string, string, error) {
	switch mode {
	case consts.ModePrimary:
		return s.filterPrimaryCandidates(candidates)
	case consts.ModeSecondary:
		return s.filterSecondaryCandidates(candidates)
	default:
		return candidates, "", nil
	}
}

func (s *Service) filterPrimaryCandidates(candidates []string) ([]string, string, error) {
	if s.BackupDBOptions.ClientCode == "" {
		return candidates, "", nil
	}

	filtered := helpers.FilterCandidatesByClientCode(candidates, s.BackupDBOptions.ClientCode)
	if len(filtered) == 0 {
		return nil, "", fmt.Errorf("tidak ada database primary dengan client code '%s' yang ditemukan", s.BackupDBOptions.ClientCode)
	}

	if len(filtered) == 1 {
		return filtered, filtered[0], nil
	}

	return filtered, "", nil
}

func (s *Service) filterSecondaryCandidates(candidates []string) ([]string, string, error) {
	if s.BackupDBOptions.ClientCode == "" && s.BackupDBOptions.Instance == "" {
		return candidates, "", nil
	}

	filtered := helpers.FilterSecondaryByClientCodeAndInstance(
		candidates,
		s.BackupDBOptions.ClientCode,
		s.BackupDBOptions.Instance,
	)

	if len(filtered) == 0 {
		if s.BackupDBOptions.ClientCode != "" && s.BackupDBOptions.Instance != "" {
			return nil, "", fmt.Errorf(
				"tidak ada database secondary dengan client code '%s' dan instance '%s' yang ditemukan",
				s.BackupDBOptions.ClientCode,
				s.BackupDBOptions.Instance,
			)
		}
		if s.BackupDBOptions.ClientCode != "" {
			return nil, "", fmt.Errorf("tidak ada database secondary dengan client code '%s' yang ditemukan", s.BackupDBOptions.ClientCode)
		}
		return nil, "", fmt.Errorf("tidak ada database secondary dengan instance '%s' yang ditemukan", s.BackupDBOptions.Instance)
	}

	if s.BackupDBOptions.Instance != "" && len(filtered) == 1 {
		return filtered, filtered[0], nil
	}

	return filtered, "", nil
}
