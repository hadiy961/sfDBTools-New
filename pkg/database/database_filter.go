package database

import (
	"context"
	"errors"
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	"strings"
)

// FilterDatabases mengambil dan memfilter daftar database dari server berdasarkan FilterOptions
// Returns: filtered database list, statistics, error
func FilterDatabases(ctx context.Context, client *Client, options types.FilterOptions) ([]string, *types.FilterStats, error) {
	// 1. Get database list from server
	allDatabases, err := client.GetDatabaseList(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mengambil daftar database: %w", err)
	}

	// Initialize stats
	stats := &types.FilterStats{
		TotalFound: len(allDatabases),
	}

	// 2. Load whitelist from file if specified (priority tertinggi)
	var whitelistFromFile []string
	if options.IncludeFile != "" {
		whitelistFromFile, err = fsops.ReadLinesFromFile(options.IncludeFile)
		if err != nil {
			return nil, nil, fmt.Errorf("gagal membaca file whitelist %s: %w", options.IncludeFile, err)
		}
		// Clean whitelist
		whitelistFromFile = helper.ListTrimNonEmpty(whitelistFromFile)
	}

	// 3. Merge whitelist: file takes priority, then IncludeDatabases
	var whitelist []string
	if len(whitelistFromFile) > 0 {
		whitelist = whitelistFromFile
	} else if len(options.IncludeDatabases) > 0 {
		whitelist = helper.ListTrimNonEmpty(options.IncludeDatabases)
	}

	// 4. Clean blacklist
	blacklist := helper.ListTrimNonEmpty(options.ExcludeDatabases)

	// 5. Filter databases
	filtered := make([]string, 0, len(allDatabases))
	for _, dbName := range allDatabases {
		dbName = strings.TrimSpace(dbName)

		// Check exclusion
		if shouldExcludeDatabase(dbName, whitelist, blacklist, options.ExcludeSystem, stats) {
			stats.TotalExcluded++
			continue
		}

		filtered = append(filtered, dbName)
	}

	stats.TotalIncluded = len(filtered)

	// Validate result
	if len(filtered) == 0 {
		return nil, stats, fmt.Errorf("database %s tidak ada (found: %d, excluded: %d)",
			strings.Join(whitelist, ", "), stats.TotalFound, stats.TotalExcluded)
	}

	return filtered, stats, nil
}

// getDatabaseList mendapatkan daftar database dari server, menerapkan filter exclude jika ada.
func (s *Client) GetDatabaseList(ctx context.Context) ([]string, error) {
	var databases []string

	rows, err := s.db.QueryContext(ctx, "SET STATEMENT max_statement_time=0 FOR SHOW DATABASES")
	if err != nil {
		return nil, errors.New("gagal mendapatkan daftar database: " + err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, errors.New("gagal membaca nama database: " + err.Error())
		}

		databases = append(databases, dbName)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New("terjadi kesalahan saat membaca daftar database: " + err.Error())
	}

	return databases, nil
}

// cleanDatabaseList membersihkan list database dari whitespace dan entry kosong
// moved to helper.ListTrimNonEmpty

// shouldExcludeDatabase menentukan apakah database harus di-exclude
// Returns true jika database harus di-exclude, false jika harus di-include
func shouldExcludeDatabase(dbName string, whitelist, blacklist []string, excludeSystem bool, stats *types.FilterStats) bool {
	// 1. Exclude empty names
	if dbName == "" {
		stats.ExcludedEmpty++
		return true
	}

	// 2. Whitelist has highest priority - if specified, only include databases in whitelist
	if len(whitelist) > 0 {
		if !helper.StringSliceContainsFold(whitelist, dbName) {
			stats.ExcludedByFile++
			return true
		}
		return false // Database is in whitelist, include it (skip other checks)
	}

	// 3. Check blacklist
	if helper.StringSliceContainsFold(blacklist, dbName) {
		stats.ExcludedByList++
		return true
	}

	// 4. Check system databases
	if excludeSystem && isSystemDatabase(dbName) {
		stats.ExcludedSystem++
		return true
	}

	// Database passed all filters
	return false
}

// isSystemDatabase memeriksa apakah database adalah system database
func isSystemDatabase(dbName string) bool {
	_, exists := types.SystemDatabases[strings.ToLower(dbName)]
	return exists
}

// containsDatabase memeriksa apakah database ada dalam list
// moved to helper.StringSliceContainsFold
