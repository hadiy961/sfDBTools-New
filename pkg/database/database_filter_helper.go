// File : pkg/database/database_filter_helper.go
// Deskripsi : Helper functions untuk database filtering dari berbagai option sources
// Author : Hadiyatna Muflihun
// Tanggal : 11 November 2025
// Last Modified : 11 November 2025

package database

import (
	"context"
	"sfDBTools/internal/types"
)

// FilterFromBackupOptions membuat FilterOptions dari BackupDBOptions dan execute filtering
// Helper untuk menghindari duplikasi di backup_filter.go
func FilterFromBackupOptions(ctx context.Context, client *Client, opts *types.BackupDBOptions) ([]string, *types.DatabaseFilterStats, error) {
	filterOpts := types.FilterOptions{
		ExcludeSystem:    opts.Filter.ExcludeSystem,
		ExcludeDatabases: opts.Filter.ExcludeDatabases,
		ExcludeDBFile:    opts.Filter.ExcludeDBFile,
		IncludeDatabases: opts.Filter.IncludeDatabases,
		IncludeFile:      opts.Filter.IncludeFile,
	}

	return FilterDatabases(ctx, client, filterOpts)
}

// FilterFromScanOptions membuat FilterOptions dari ScanOptions dan execute filtering
// Helper untuk menghindari duplikasi di dbscan_filter.go
func FilterFromScanOptions(ctx context.Context, client *Client, opts *types.ScanOptions) ([]string, *types.DatabaseFilterStats, error) {
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

	return FilterDatabases(ctx, client, filterOpts)
}
