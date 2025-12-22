package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
)

// GetFilteredDatabases mengambil dan memfilter daftar database sesuai aturan
// Untuk command filter tanpa include/exclude flags: tampilkan multi-select
// Untuk command all atau filter dengan flags: gunakan filter biasa
func (s *Service) GetFilteredDatabases(ctx context.Context, client *database.Client) ([]string, *types.FilterStats, error) {
	hasIncludeFlags := len(s.BackupDBOptions.Filter.IncludeDatabases) > 0 || s.BackupDBOptions.Filter.IncludeFile != ""

	// Jika ada include flags, selalu gunakan filter biasa (tidak perlu multi-select)
	if hasIncludeFlags {
		return database.FilterFromBackupOptions(ctx, client, s.BackupDBOptions)
	}

	// Jika ini command filter (IsFilterCommand=true) dan tidak ada include/exclude yang di-set manual
	// maka tampilkan multi-select
	isFilterMode := s.BackupDBOptions.Filter.IsFilterCommand

	// Untuk command filter tanpa include dan exclude manual → multi-select
	if isFilterMode {
		return s.getFilteredDatabasesWithMultiSelect(ctx, client)
	}

	// Untuk command all atau filter dengan flags → gunakan filter biasa dengan nilai default
	return database.FilterFromBackupOptions(ctx, client, s.BackupDBOptions)
}

// displayFilterWarnings menampilkan warning messages untuk filter stats
func (s *Service) displayFilterWarnings(stats *types.FilterStats) {
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
