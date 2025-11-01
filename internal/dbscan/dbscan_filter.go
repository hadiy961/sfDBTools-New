package dbscan

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
		ExcludeSystem:    s.ScanOptions.ExcludeSystem,
		ExcludeDatabases: s.ScanOptions.ExcludeList,
		IncludeDatabases: s.ScanOptions.IncludeList,
	}

	// Untuk mode single, gunakan SourceDatabase yang telah ditentukan
	if s.ScanOptions.Mode == "single" && s.ScanOptions.SourceDatabase != "" {
		filterOpts.IncludeDatabases = []string{s.ScanOptions.SourceDatabase}
		// Untuk single database, tidak perlu exclude system databases
		filterOpts.ExcludeSystem = false
	}

	// Untuk mode database, gunakan file list jika tersedia
	if s.ScanOptions.Mode == "database" && s.ScanOptions.DatabaseList.File != "" {
		filterOpts.IncludeFile = s.ScanOptions.DatabaseList.File
	}

	// Execute filtering
	filtered, stats, err := database.FilterDatabases(ctx, client, filterOpts)
	if err != nil {
		return nil, nil, err
	}

	// Convert FilterStats to DatabaseFilterStats (untuk compatibility dengan existing code)
	dbStats := &types.DatabaseFilterStats{
		TotalFound:     stats.TotalFound,
		ToScan:         stats.TotalIncluded,
		ExcludedSystem: stats.ExcludedSystem,
		ExcludedByList: stats.ExcludedByList,
		ExcludedByFile: stats.ExcludedByFile,
		ExcludedEmpty:  stats.ExcludedEmpty,
	}

	return filtered, dbStats, nil
}
