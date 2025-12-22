package helpers

import (
	"context"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
)

// FilterFromScanOptions membuat FilterOptions dari ScanOptions dan menjalankan filtering.
func FilterFromScanOptions(ctx context.Context, client *database.Client, opts *types.ScanOptions) ([]string, *types.FilterStats, error) {
	filterOpts := types.FilterOptions{
		ExcludeSystem:    opts.ExcludeSystem,
		ExcludeDatabases: opts.ExcludeList,
		ExcludeDBFile:    opts.ExcludeFile,
		IncludeDatabases: opts.IncludeList,
	}

	// Untuk mode single, gunakan SourceDatabase yang telah ditentukan
	if opts.Mode == "single" && opts.SourceDatabase != "" {
		filterOpts.IncludeDatabases = []string{opts.SourceDatabase}
		// Untuk single database, tidak perlu exclude system databases
		filterOpts.ExcludeSystem = false
	}

	// Untuk mode database, gunakan file list jika tersedia
	if opts.Mode == "database" && opts.DatabaseList.File != "" {
		filterOpts.IncludeFile = opts.DatabaseList.File
	}

	return database.FilterDatabases(ctx, client, filterOpts)
}
