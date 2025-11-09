// File : internal/restore/restore_pattern.go
// Deskripsi : Pattern-based database name extraction dari filename backup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package restore

import (
	"path/filepath"
	"regexp"
	"strings"
)

// extractDatabaseNameFromPattern extract nama database dari filename berdasarkan name_pattern config
// Strategy: Convert pattern ke regex dengan multiple capture groups,
// kemudian identify mana yang {database}
func extractDatabaseNameFromPattern(filename, namePattern string) string {
	base := filepath.Base(filename)

	// Remove extensions
	base = strings.TrimSuffix(base, ".enc")
	base = strings.TrimSuffix(base, ".gz")
	base = strings.TrimSuffix(base, ".zst")
	base = strings.TrimSuffix(base, ".xz")
	base = strings.TrimSuffix(base, ".zlib")
	base = strings.TrimSuffix(base, ".sql")

	// Jika pattern kosong atau tidak ada {database}, fallback ke old logic
	if namePattern == "" || !strings.Contains(namePattern, "{database}") {
		return extractDatabaseNameLegacy(base)
	}

	// Build regex dan track which capture group is database
	regexPattern, dbGroupIndex := convertPatternToRegexWithGroups(namePattern)

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		// Jika regex error, fallback ke legacy
		return extractDatabaseNameLegacy(base)
	}

	matches := re.FindStringSubmatch(base)
	if len(matches) > dbGroupIndex {
		return matches[dbGroupIndex]
	}

	// Fallback jika regex tidak match
	return extractDatabaseNameLegacy(base)
}

// convertPatternToRegexWithGroups convert pattern ke regex dan return group index untuk database
// Return: (regexPattern, databaseGroupIndex)
func convertPatternToRegexWithGroups(namePattern string) (string, int) {
	pattern := regexp.QuoteMeta(namePattern)
	dbGroupIndex := 0
	currentGroupIndex := 0

	// Define all placeholders dan regex-nya
	// Placeholders yang jadi capture group akan increment group index
	placeholders := []struct {
		name      string
		regex     string
		isCapture bool
	}{
		{`\{database\}`, `(.+?)`, true},         // Capture group - this is what we want
		{`\{client_code\}`, `(.+?)`, true},      // Capture group
		{`\{hostname\}`, `([^\s]+)`, true},      // Capture group
		{`\{timestamp\}`, `\d{8}_\d{6}`, false}, // Not captured
		{`\{year\}`, `\d{4}`, false},
		{`\{month\}`, `\d{2}`, false},
		{`\{day\}`, `\d{2}`, false},
		{`\{hour\}`, `\d{2}`, false},
		{`\{minute\}`, `\d{2}`, false},
		{`\{second\}`, `\d{2}`, false},
	}

	// Track mana placeholder {database} di pattern
	for _, ph := range placeholders {
		if strings.Contains(pattern, ph.name) {
			if ph.isCapture {
				currentGroupIndex++
			}
			if ph.name == `\{database\}` {
				dbGroupIndex = currentGroupIndex
			}
			pattern = strings.ReplaceAll(pattern, ph.name, ph.regex)
		}
	}

	return "^" + pattern + "$", dbGroupIndex
}

// extractDatabaseNameLegacy adalah fallback function dengan logic lama
// Mencari pattern 8 digits untuk mendeteksi date
// Return empty string jika tidak confident untuk trigger interactive prompt
func extractDatabaseNameLegacy(base string) string {
	parts := strings.Split(base, "_")

	// Find the timestamp part (8 digits for date)
	for i := 0; i < len(parts); i++ {
		// Check if this part is a date (8 digits)
		if len(parts[i]) == 8 && isAllDigits(parts[i]) {
			// The database name is everything before this index
			if i > 0 {
				return strings.Join(parts[:i], "_")
			}
			// Jika date di index 0, tidak valid
			return ""
		}
	}

	// Tidak ketemu pattern timestamp yang jelas - return empty untuk trigger prompt
	// Ini lebih aman daripada guess yang mungkin salah
	return ""
} // isAllDigits check apakah string hanya berisi digit
func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
