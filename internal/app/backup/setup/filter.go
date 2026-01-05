package setup

import (
	"context"
	"fmt"

	"sfDBTools/internal/app/backup/selection"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
)

// GetFilteredDatabases fetches and filters DB list.
// For 'filter' command without include flags, it shows a multi-select.
func (s *Setup) GetFilteredDatabases(ctx context.Context, client *database.Client) ([]string, *types.FilterStats, error) {
	hasIncludeFlags := len(s.Options.Filter.IncludeDatabases) > 0 || s.Options.Filter.IncludeFile != ""
	if hasIncludeFlags {
		return filterFromBackupOptions(ctx, client, s.Options)
	}

	isFilterMode := s.Options.Filter.IsFilterCommand
	if isFilterMode {
		return selection.New(s.Log, s.Options).GetFilteredDatabasesWithMultiSelect(ctx, client)
	}

	return filterFromBackupOptions(ctx, client, s.Options)
}

func filterFromBackupOptions(ctx context.Context, client *database.Client, opts *types_backup.BackupDBOptions) ([]string, *types.FilterStats, error) {
	filterOpts := types.FilterOptions{
		ExcludeSystem:    opts.Filter.ExcludeSystem,
		ExcludeDatabases: opts.Filter.ExcludeDatabases,
		ExcludeDBFile:    opts.Filter.ExcludeDBFile,
		IncludeDatabases: opts.Filter.IncludeDatabases,
		IncludeFile:      opts.Filter.IncludeFile,
	}
	dbFiltered, stats, err := database.FilterDatabases(ctx, client, filterOpts)
	if err != nil {
		return nil, stats, err
	}

	return dbFiltered, stats, nil
}

func (s *Setup) DisplayFilterWarnings(stats *types.FilterStats) {
	ui.PrintWarning("Kemungkinan penyebab:")

	if stats.TotalExcluded == stats.TotalFound {
		ui.PrintWarning(fmt.Sprintf("  • Semua database (%d) dikecualikan oleh filter exclude", stats.TotalExcluded))
	}

	if len(stats.NotFoundInInclude) > 0 {
		ui.PrintWarning("  • Database yang diminta di include list tidak ditemukan:")
		for _, db := range stats.NotFoundInInclude {
			ui.PrintWarning(fmt.Sprintf("    - %s", db))
		}
	}

	if len(stats.NotFoundInWhitelist) > 0 {
		ui.PrintWarning("  • Database dari whitelist file tidak ditemukan:")
		for _, db := range stats.NotFoundInWhitelist {
			ui.PrintWarning(fmt.Sprintf("    - %s", db))
		}
	}

	if len(stats.NotFoundInExclude) > 0 {
		ui.PrintWarning("  • Database yang diminta di exclude list tidak ditemukan:")
		for _, db := range stats.NotFoundInExclude {
			ui.PrintWarning(fmt.Sprintf("    - %s", db))
		}
	}

	if len(stats.NotFoundInBlacklist) > 0 {
		ui.PrintWarning("  • Database dari exclude file tidak ditemukan:")
		for _, db := range stats.NotFoundInBlacklist {
			ui.PrintWarning(fmt.Sprintf("    - %s", db))
		}
	}
}
