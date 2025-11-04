package backup

import (
	"context"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
)

// GetFilteredDatabases mengambil dan memfilter daftar database sesuai aturan.
// Menggunakan general database filtering system dari pkg/database
func (s *Service) GetFilteredDatabases(ctx context.Context, client *database.Client) ([]string, *types.DatabaseFilterStats, error) {

	// Setup filter options from ScanOptions. If a whitelist file is enabled, include it.
	filterOpts := types.FilterOptions{
		ExcludeSystem:    s.BackupDBOptions.Filter.ExcludeSystem,
		ExcludeDatabases: s.BackupDBOptions.Filter.ExcludeDatabases,
		ExcludeDBFile:    s.BackupDBOptions.Filter.ExcludeDBFile,
		IncludeDatabases: s.BackupDBOptions.Filter.IncludeDatabases,
		IncludeFile:      s.BackupDBOptions.Filter.IncludeFile,
	}

	// Execute filtering
	filtered, stats, err := database.FilterDatabases(ctx, client, filterOpts)
	if err != nil {
		return nil, nil, err
	}

	// Return directly - DatabaseFilterStats is now an alias for FilterStats
	return filtered, stats, nil
}
