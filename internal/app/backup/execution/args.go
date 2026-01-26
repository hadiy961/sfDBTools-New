// File : internal/app/backup/execution/args.go
// Deskripsi : Mysqldump arguments builder dan password masking
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2026-01-26

package execution

import (
	"strconv"
	"strings"

	"sfdbtools/internal/domain"
)

// BuildMysqldumpArgs membuat argumen mysqldump dari konfigurasi backup.
// Function ini pure logic tanpa wrapper - langsung menggunakan types yang sudah ada.
func BuildMysqldumpArgs(
	baseDumpArgs string,
	dbInfo domain.DBInfo,
	filter domain.FilterOptions,
	dbFiltered []string,
	singleDB string,
	totalDBFound int,
	skipTablesData []string,
) []string {
	var args []string

	// Connection parameters
	if dbInfo.Host != "" {
		args = append(args, "--host="+dbInfo.Host)
	}
	if dbInfo.Port != 0 {
		args = append(args, "--port="+strconv.Itoa(dbInfo.Port))
	}
	if dbInfo.User != "" {
		args = append(args, "--user="+dbInfo.User)
	}
	if dbInfo.Password != "" {
		args = append(args, "--password="+dbInfo.Password)
	}

	// Base mysqldump args dari config
	if baseDumpArgs != "" {
		args = append(args, strings.Fields(baseDumpArgs)...)
	}

	// Data filter
	if filter.ExcludeData {
		args = append(args, "--no-data")
	}

	// Skip specific tables data (hanya backup struktur untuk table tertentu)
	// Untuk mode primary/secondary: tsflpageview, tsfltokensess, tcllusersession
	if len(skipTablesData) > 0 {
		// Tentukan database yang akan digunakan untuk --ignore-table-data
		var targetDB string
		if singleDB != "" {
			targetDB = singleDB
		} else if len(dbFiltered) == 1 {
			targetDB = dbFiltered[0]
		}

		// Hanya tambahkan --ignore-table-data jika kita backup single database
		// Untuk multi-database backup, skip fitur ini karena table name bisa konflik
		if targetDB != "" {
			for _, table := range skipTablesData {
				if table != "" {
					args = append(args, "--ignore-table-data="+targetDB+"."+table)
				}
			}
		}
	}

	// CASE 1: Single database eksplisit
	if singleDB != "" {
		args = append(args, singleDB)
		return args
	}

	// Check apakah ada filter aktif
	hasFilter := len(filter.ExcludeDatabases) > 0 ||
		len(filter.IncludeDatabases) > 0 ||
		filter.ExcludeSystem ||
		filter.ExcludeDBFile != "" ||
		filter.IncludeFile != ""

	// CASE 2: Full backup (no filter, all databases)
	if !hasFilter && len(dbFiltered) == totalDBFound {
		args = append(args, "--all-databases")
		return args
	}

	// CASE 3: Filtered databases
	if len(dbFiltered) == 1 {
		// Single database dari hasil filter
		args = append(args, dbFiltered[0])
		return args
	}
	if len(dbFiltered) > 1 {
		// Multiple databases
		args = append(args, "--databases")
		args = append(args, dbFiltered...)
	}

	return args
}
