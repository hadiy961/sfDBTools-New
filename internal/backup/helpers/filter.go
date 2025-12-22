// File : internal/backup/helpers/filter.go
// Deskripsi : Helper function untuk database filtering dari BackupDBOptions
// Author : Hadiyatna Muflihun
// Tanggal : 22 Desember 2024
// Last Modified : 22 Desember 2024
// Moved from: pkg/database/database_filter_helper.go (partial)

package helpers

import (
	"context"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/database"
)

// FilterFromBackupOptions membuat FilterOptions dari BackupDBOptions dan execute filtering
// Helper untuk menghindari duplikasi di backup_filter.go
func FilterFromBackupOptions(ctx context.Context, client *database.Client, opts *types_backup.BackupDBOptions) ([]string, *types.FilterStats, error) {
	filterOpts := types.FilterOptions{
		ExcludeSystem:    opts.Filter.ExcludeSystem,
		ExcludeDatabases: opts.Filter.ExcludeDatabases,
		ExcludeDBFile:    opts.Filter.ExcludeDBFile,
		IncludeDatabases: opts.Filter.IncludeDatabases,
		IncludeFile:      opts.Filter.IncludeFile,
	}

	return database.FilterDatabases(ctx, client, filterOpts)
}
