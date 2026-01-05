package database

import (
	"context"

	dbscanmodel "sfDBTools/internal/app/dbscan/model"
	"sfDBTools/internal/domain"
)

// FilterFromScanOptions membuat FilterOptions dari ScanOptions dan execute filtering.
//
// Deprecated: gunakan internal/dbscan/helpers.FilterFromScanOptions.
func FilterFromScanOptions(ctx context.Context, client *Client, opts *dbscanmodel.ScanOptions) ([]string, *domain.FilterStats, error) {
	filterOpts := domain.FilterOptions{
		ExcludeSystem:    opts.ExcludeSystem,
		ExcludeDatabases: opts.ExcludeList,
		ExcludeDBFile:    opts.ExcludeFile,
		IncludeDatabases: opts.IncludeList,
	}

	if opts.Mode == "single" && opts.SourceDatabase != "" {
		filterOpts.IncludeDatabases = []string{opts.SourceDatabase}
		filterOpts.ExcludeSystem = false
	}

	if opts.Mode == "database" && opts.DatabaseList.File != "" {
		filterOpts.IncludeFile = opts.DatabaseList.File
	}

	return FilterDatabases(ctx, client, filterOpts)
}
