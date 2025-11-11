// File : internal/restore/restore_pattern.go
// Deskripsi : Pattern-based database name extraction dari filename backup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-10
//
// Fixed Pattern:
// ==============
// Pattern yang digunakan adalah FIXED dan tidak bisa dikonfigurasi:
// {database}_{year}{month}{day}_{hour}{minute}{second}_{hostname}
//
// Contoh filename:
// - mydb_20251110_143025_localhost.sql.gz.enc
// - testdb_20251110_143025_dbserver.sql.zst
//
// Extraction algorithm:
// 1. Remove semua ekstensi (.sql, .gz, .zst, .xz, .enc, dll)
// 2. Parse menggunakan regex untuk fixed pattern
// 3. Extract database name dari capture group pertama
// 4. Return empty string jika tidak match (akan trigger prompt ke user)

package restore

import (
	"regexp"
	"sfDBTools/pkg/helper"
)

// Fixed pattern yang digunakan untuk filename backup
const FixedBackupPattern = "{database}_{year}{month}{day}_{hour}{minute}{second}_{hostname}"

// extractDatabaseNameFromPattern extract nama database dari filename menggunakan fixed pattern
// Return empty string jika filename tidak sesuai pattern (akan trigger interactive prompt)
func extractDatabaseNameFromPattern(filename string) string {
	// Remove all backup extensions menggunakan helper
	base := helper.StripAllBackupExtensions(filename)

	// Build regex untuk fixed pattern: {database}_{year}{month}{day}_{hour}{minute}{second}_{hostname}
	// Pattern: (.+?)_\d{8}_\d{6}_([^_]+)
	// Group 1: database name (minimal match)
	// Group 2: hostname
	regexPattern := `^(.+?)_\d{8}_\d{6}_([^_]+)$`

	re := regexp.MustCompile(regexPattern)
	matches := re.FindStringSubmatch(base)

	if len(matches) > 1 {
		// Return database name (group 1)
		return matches[1]
	}

	// Tidak match dengan pattern, return empty string
	// Ini akan trigger interactive prompt di restore_single.go
	return ""
}
