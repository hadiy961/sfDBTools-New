package setup

import (
	"context"
	"fmt"

	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/app/backup/selection"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/ui/print"
)

// GetFilteredDatabases fetches and filters DB list.
// For 'filter' command without include flags, it shows a multi-select.
func (s *Setup) GetFilteredDatabases(ctx context.Context, client *database.Client) ([]string, *domain.FilterStats, error) {
	hasIncludeFlags := len(s.Options.Filter.IncludeDatabases) > 0 || s.Options.Filter.IncludeFile != ""
	if hasIncludeFlags {
		return filterFromBackupOptions(ctx, client, s.Options)
	}

	isFilterMode := s.Options.Filter.IsFilterCommand
	if isFilterMode {
		selectedDBs, stats, err := selection.New(s.Log, s.Options).GetFilteredDatabasesWithMultiSelect(ctx, client)
		if err != nil {
			return nil, stats, err
		}

		// Persist pilihan sebagai include list agar flow tidak meminta multi-select lagi
		// ketika user mengubah opsi (mis: filename/encryption/compression) dan session loop re-run.
		s.Options.Filter.IncludeDatabases = selectedDBs
		s.Options.Filter.IncludeFile = ""

		return selectedDBs, stats, nil
	}

	return filterFromBackupOptions(ctx, client, s.Options)
}

func filterFromBackupOptions(ctx context.Context, client *database.Client, opts *types_backup.BackupDBOptions) ([]string, *domain.FilterStats, error) {
	filterOpts := domain.FilterOptions{
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

	// Mode primary/secondary membutuhkan filtering tambahan berbasis konvensi penamaan.
	// Ini harus dilakukan TANPA prompt, supaya mode daemon/background tetap bisa berjalan.
	switch opts.Mode {
	case consts.ModePrimary, consts.ModeSecondary, consts.ModeSingle:
		dbFiltered = selection.FilterCandidatesByMode(dbFiltered, opts.Mode)
	}
	if opts.Mode == consts.ModePrimary {
		// Optional: jika client code diisi, sempitkan kandidat.
		dbFiltered = selection.FilterCandidatesByClientCode(dbFiltered, opts.ClientCode)
	}
	if opts.Mode == consts.ModeSecondary {
		// Optional: client code dan/atau instance bisa dipakai.
		if opts.ClientCode != "" || opts.Instance != "" {
			dbFiltered = selection.FilterSecondaryByClientCodeAndInstance(dbFiltered, opts.ClientCode, opts.Instance)
		}
	}

	return dbFiltered, stats, nil
}

func (s *Setup) DisplayFilterWarnings(stats *domain.FilterStats) {
	print.PrintWarning("Kemungkinan penyebab:")

	if stats.TotalExcluded == stats.TotalFound {
		print.PrintWarning(fmt.Sprintf("  • Semua database (%d) dikecualikan oleh filter exclude", stats.TotalExcluded))
	}

	if len(stats.NotFoundInInclude) > 0 {
		print.PrintWarning("  • Database yang diminta di include list tidak ditemukan:")
		for _, db := range stats.NotFoundInInclude {
			print.PrintWarning(fmt.Sprintf("    - %s", db))
		}
	}

	if len(stats.NotFoundInWhitelist) > 0 {
		print.PrintWarning("  • Database dari whitelist file tidak ditemukan:")
		for _, db := range stats.NotFoundInWhitelist {
			print.PrintWarning(fmt.Sprintf("    - %s", db))
		}
	}

	if len(stats.NotFoundInExclude) > 0 {
		print.PrintWarning("  • Database yang diminta di exclude list tidak ditemukan:")
		for _, db := range stats.NotFoundInExclude {
			print.PrintWarning(fmt.Sprintf("    - %s", db))
		}
	}

	if len(stats.NotFoundInBlacklist) > 0 {
		print.PrintWarning("  • Database dari exclude file tidak ditemukan:")
		for _, db := range stats.NotFoundInBlacklist {
			print.PrintWarning(fmt.Sprintf("    - %s", db))
		}
	}
}
