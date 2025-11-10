// File : internal/restore/restore_pattern.go
// Deskripsi : Pattern-based database name extraction dari filename backup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-10
//
// Pattern Support:
// ================
// Supported placeholders:
// - {database}      : Nama database (extract target)
// - {timestamp}     : Format YYYYMMDD_HHMMSS
// - {year}          : YYYY
// - {month}         : MM
// - {day}           : DD
// - {hour}          : HH
// - {minute}        : MM
// - {second}        : SS
// - {hostname}      : Server hostname
// - {client_code}   : Client identifier
//
// Pattern dapat dalam urutan apapun dengan constraints:
// ✓ RECOMMENDED PATTERNS:
// - {database}_{year}{month}{day}_{hour}{minute}{second}_{hostname}
// - {timestamp}_{client_code}_{database}_{hostname}
// - {year}{month}{day}_{hour}{minute}{second}_{hostname}_{database}  (database di akhir)
// - {database}_{timestamp}_{hostname}
//
// ⚠ PROBLEMATIC PATTERNS (ambiguous, tidak recommended):
// - {year}-{month}-{day}_{hour}-{database}{minute}{second}_{hostname}_{client_code}
//   Issue: Database surrounded by numeric placeholders, regex ambiguity
//   Solution: Use clear separators, database dibatasi dengan _ atau -
//
// BEST PRACTICES:
// 1. Gunakan underscore (_) sebagai primary separator
// 2. Hyphen (-) ok untuk date components saja
// 3. Database name jangan langsung adjacent dengan numeric placeholders
// 4. Gunakan clear delimiters: {database}_{next_placeholder}
// 5. Avoid mixing separators dalam placeholder sequence
//
// Extraction algorithm:
// 1. Parse pattern untuk identify semua placeholder
// 2. Convert ke regex pattern dengan tracking capture groups
// 3. Match filename terhadap regex dengan backtracking
// 4. Extract database name dari capture group yang sesuai
// 5. Fallback ke legacy heuristic jika pattern match gagal

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

	// Validate pattern (warning only, tidak fail)
	validatePatternFormat(namePattern)

	// Build regex dan track which capture group is database
	regexPattern, dbGroupIndex := convertPatternToRegexWithGroups(namePattern)

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		// Jika regex error, fallback ke legacy
		return extractDatabaseNameLegacy(base)
	}

	matches := re.FindStringSubmatch(base)
	if len(matches) > dbGroupIndex && dbGroupIndex > 0 {
		return matches[dbGroupIndex]
	}

	// Fallback jika regex tidak match
	return extractDatabaseNameLegacy(base)
}

// validatePatternFormat validate pattern format dan warn jika ada issues
// Tidak fail, hanya log warning untuk best practices
func validatePatternFormat(namePattern string) {
	issues := []string{}

	// Check 1: Database surrounded by numeric patterns (ambiguous)
	if strings.Contains(namePattern, "{hour}{database}{minute}") ||
		strings.Contains(namePattern, "{minute}{database}{second}") ||
		strings.Contains(namePattern, "{month}{database}{day}") {
		issues = append(issues, "database surrounded by numeric patterns - may cause extraction ambiguity")
	}

	// Check 2: Database without clear separator
	if strings.Contains(namePattern, "{database}{year}") ||
		strings.Contains(namePattern, "{database}{month}") ||
		strings.Contains(namePattern, "{database}{day}") ||
		strings.Contains(namePattern, "{database}{hour}") {
		issues = append(issues, "database not separated from numeric components")
	}

	// Check 3: Multiple separators mixed (hyphen + underscore in way that confuses)
	dashCount := strings.Count(namePattern, "-")
	underscoreCount := strings.Count(namePattern, "_")
	if dashCount > 3 && underscoreCount > 0 {
		issues = append(issues, "complex mix of separators - recommended to use consistent separator")
	}

	// Log warnings jika ada issues (bukan error, hanya warning)
	if len(issues) > 0 {
		// Bisa log di sini jika ada logger, untuk sekarang just comment
		// untuk testing purposes
		_ = issues
	}
}

// convertPatternToRegexWithGroups convert pattern ke regex dan return group index untuk database
// Support pattern apapun: {timestamp}_{client_code}_{database}_{hostname} bisa!
// Return: (regexPattern, databaseGroupIndex)
//
// Contoh patterns yang supported:
// - '{database}_{year}{month}{day}_{hour}{minute}{second}_{hostname}'
// - '{timestamp}_{client_code}_{database}_{hostname}'
// - '{database}_{timestamp}_{hostname}'
// - dsb
func convertPatternToRegexWithGroups(namePattern string) (string, int) {
	// Mapping placeholder ke regex pattern dan apakah capture group
	placeholderMap := map[string]struct {
		regex     string
		isCapture bool
	}{
		"{database}":    {"(.+?)", true},   // Minimal match - capture
		"{client_code}": {"(.+?)", true},   // Capture
		"{hostname}":    {"([^_]+)", true}, // Capture - tidak boleh ada underscore
		"{timestamp}":   {"\\d{8}_\\d{6}", false},
		"{year}":        {"\\d{4}", false},
		"{month}":       {"\\d{2}", false},
		"{day}":         {"\\d{2}", false},
		"{hour}":        {"\\d{2}", false},
		"{minute}":      {"\\d{2}", false},
		"{second}":      {"\\d{2}", false},
	}

	// Parse pattern dan build regex dengan tracking group index
	pattern := namePattern
	dbGroupIndex := 0
	currentGroupIndex := 0

	// Strategy: iterate through placeholders dalam pattern untuk maintain order
	// dan track capture groups dengan benar

	// Buat list placeholder yang ada di pattern, dalam urutan kemunculannya
	type placeholderInfo struct {
		placeholder string
		position    int
		value       struct {
			regex     string
			isCapture bool
		}
	}

	var placeholdersList []placeholderInfo

	// Find semua placeholder dalam pattern
	for placeholder, info := range placeholderMap {
		if strings.Contains(pattern, placeholder) {
			pos := strings.Index(pattern, placeholder)
			placeholdersList = append(placeholdersList, placeholderInfo{
				placeholder: placeholder,
				position:    pos,
				value:       info,
			})
		}
	}

	// Sort by position untuk maintain order
	for i := 0; i < len(placeholdersList)-1; i++ {
		for j := i + 1; j < len(placeholdersList); j++ {
			if placeholdersList[i].position > placeholdersList[j].position {
				placeholdersList[i], placeholdersList[j] = placeholdersList[j], placeholdersList[i]
			}
		}
	}

	// Replace semua placeholder dengan regex, track group indices
	for _, ph := range placeholdersList {
		if ph.value.isCapture {
			currentGroupIndex++
		}

		if ph.placeholder == "{database}" {
			dbGroupIndex = currentGroupIndex
		}

		pattern = strings.ReplaceAll(pattern, ph.placeholder, ph.value.regex)
	}

	// Escape remaining special regex chars (underscore dan literal characters)
	// yang belum di-replace
	pattern = escapeRemainingSpecials(pattern)

	return "^" + pattern + "$", dbGroupIndex
}

// escapeRemainingSpecials escape special regex chars yang belum di-escape
func escapeRemainingSpecials(s string) string {
	// Hanya escape chars yang special untuk regex dan bukan capture group
	replacements := map[string]string{
		".": "\\.",
		"+": "\\+",
		"*": "\\*",
		"?": "\\?",
		"[": "\\[",
		"]": "\\]",
		"|": "\\|",
	}

	result := s
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	return result
} // extractDatabaseNameLegacy adalah fallback ketika pattern matching gagal
// Strategy simple: cari database name sebelum timestamp numeric pattern
// Pattern standar: {database}_{timestamp} -> extract bagian sebelum digits
func extractDatabaseNameLegacy(base string) string {
	// Strategi 1: Cari pattern di mana ada underscore diikuti 8 digit (YYYYMMDD)
	// Format: dbname_YYYYMMDD_... -> extract dbname
	parts := strings.Split(base, "_")

	for i := 0; i < len(parts); i++ {
		// Cek apakah part ini adalah angka 8 digit (date)
		if len(parts[i]) >= 8 && isAllDigits(parts[i][:8]) {
			// Database name adalah semua parts sebelum index ini
			if i > 0 {
				return strings.Join(parts[:i], "_")
			}
			break
		}
	}

	// Strategi 2: Jika tidak ada numeric pattern jelas, gunakan heuristic
	// Asumsikan database name adalah semua parts kecuali yang terlihat seperti:
	// - server hostname
	// - timestamp
	//
	// Untuk safety, return seluruh base jika tidak ada separator yang jelas
	// Ini cocok untuk backup files sederhana tanpa struktur standar
	if len(parts) > 0 {
		// Jika hanya 1-2 parts, return semuanya sebagai database name
		if len(parts) <= 2 {
			return base
		}

		// Cek apakah bagian terakhir terlihat seperti hostname (misal: localhost, server1, etc)
		// Jika iya, kemungkinan format adalah: db_timestamp_hostname
		// Return bagian pertama sebagai database name
		lastPart := parts[len(parts)-1]
		if len(lastPart) > 0 && !isAllDigits(lastPart) && strings.ContainsAny(lastPart, "abcdefghijklmnopqrstuvwxyz0123456789") {
			// Tampak seperti hostname
			// Cek apakah part sebelumnya adalah timestamp (8-14 digit)
			if len(parts) >= 2 {
				secondLastPart := parts[len(parts)-2]
				if len(secondLastPart) >= 8 && isAllDigits(secondLastPart) {
					// Format: db_timestamp_hostname
					return strings.Join(parts[:len(parts)-2], "_")
				}
			}
		}

		// Default: return base sebagai database name
		return base
	}

	return ""
}

// isAllDigits check apakah string hanya berisi digit
func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
