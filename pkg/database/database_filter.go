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
		TotalFound:          len(allDatabases),
		NotFoundInInclude:   []string{},
		NotFoundInExclude:   []string{},
		NotFoundInWhitelist: []string{},
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

	// 3. Load blacklist from file if specified
	var blacklistFromFile []string
	if options.ExcludeDBFile != "" {
		blacklistFromFile, err = fsops.ReadLinesFromFile(options.ExcludeDBFile)
		if err != nil {
			return nil, nil, fmt.Errorf("gagal membaca file blacklist %s: %w", options.ExcludeDBFile, err)
		}
		// Clean blacklist
		blacklistFromFile = helper.ListTrimNonEmpty(blacklistFromFile)
	}

	// 4. Merge whitelist: combine file and direct list
	var whitelist []string
	if len(whitelistFromFile) > 0 {
		whitelist = append(whitelist, whitelistFromFile...)
	}
	if len(options.IncludeDatabases) > 0 {
		whitelist = append(whitelist, helper.ListTrimNonEmpty(options.IncludeDatabases)...)
	}
	// Remove duplicates from whitelist using map
	var isFromFile bool
	if len(whitelist) > 0 {
		seen := make(map[string]bool)
		uniqueWhitelist := make([]string, 0, len(whitelist))
		for _, db := range whitelist {
			dbLower := strings.ToLower(strings.TrimSpace(db))
			if dbLower != "" && !seen[dbLower] {
				seen[dbLower] = true
				uniqueWhitelist = append(uniqueWhitelist, db)
			}
		}
		whitelist = uniqueWhitelist
		// Track if any came from file for warning purposes
		isFromFile = len(whitelistFromFile) > 0
	}

	// 5. Merge blacklist: combine file and direct list
	var blacklist []string
	var blacklistIsFromFile bool
	if len(blacklistFromFile) > 0 {
		blacklist = append(blacklist, blacklistFromFile...)
		blacklistIsFromFile = true
	}
	if len(options.ExcludeDatabases) > 0 {
		blacklist = append(blacklist, helper.ListTrimNonEmpty(options.ExcludeDatabases)...)
	}
	// Remove duplicates from blacklist using map
	if len(blacklist) > 0 {
		seen := make(map[string]bool)
		uniqueBlacklist := make([]string, 0, len(blacklist))
		for _, db := range blacklist {
			dbLower := strings.ToLower(strings.TrimSpace(db))
			if dbLower != "" && !seen[dbLower] {
				seen[dbLower] = true
				uniqueBlacklist = append(uniqueBlacklist, db)
			}
		}
		blacklist = uniqueBlacklist
	}

	// 6. Validate whitelist - check if databases in whitelist exist on server
	if len(whitelist) > 0 {
		for _, dbName := range whitelist {
			if !helper.StringSliceContainsFold(allDatabases, dbName) {
				if isFromFile {
					stats.NotFoundInWhitelist = append(stats.NotFoundInWhitelist, dbName)
				} else {
					stats.NotFoundInInclude = append(stats.NotFoundInInclude, dbName)
				}
			}
		}
	}

	// 6. Validate blacklist - check if databases in blacklist exist on server
	if len(blacklist) > 0 {
		for _, dbName := range blacklist {
			if !helper.StringSliceContainsFold(allDatabases, dbName) {
				if blacklistIsFromFile && helper.StringSliceContainsFold(blacklistFromFile, dbName) {
					stats.NotFoundInBlacklist = append(stats.NotFoundInBlacklist, dbName)
				} else {
					stats.NotFoundInExclude = append(stats.NotFoundInExclude, dbName)
				}
			}
		}
	}

	// 7. Filter databases
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

	// Return hasil tanpa error - biar caller yang handle empty result dengan UI yang lebih baik
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
	if excludeSystem && IsSystemDatabase(dbName) {
		stats.ExcludedSystem++
		return true
	}

	// Database passed all filters
	return false
}

// IsSystemDatabase memeriksa apakah database adalah system database
// Function ini di-export agar bisa digunakan di package lain
func IsSystemDatabase(dbName string) bool {
	_, exists := types.SystemDatabases[strings.ToLower(dbName)]
	return exists
}

// containsDatabase memeriksa apakah database ada dalam list
// moved to helper.StringSliceContainsFold
