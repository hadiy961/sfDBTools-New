// File : internal/app/profile/helpers/common/string_utils.go
// Deskripsi : Common string utilities (untuk eliminate duplikasi)
// Author : Hadiyatna Muflihun
// Tanggal : 25 Januari 2026
// Last Modified : 25 Januari 2026

package common

import (
	"strings"

	profileconn "sfdbtools/internal/app/profile/connection"
)

// Trim returns trimmed string atau empty string jika nil/whitespace
func Trim(s string) string {
	return strings.TrimSpace(s)
}

// TrimLower returns trimmed lowercase string
func TrimLower(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// IsEmpty checks if string kosong setelah trim
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty checks if string tidak kosong setelah trim
func IsNotEmpty(s string) bool {
	return strings.TrimSpace(s) != ""
}

// NormalizeName normalizes profile name (trim + remove suffix)
func NormalizeName(name string) string {
	return profileconn.TrimProfileSuffix(strings.TrimSpace(name))
}

// SafeString returns non-empty string atau fallback
func SafeString(s, fallback string) string {
	if v := strings.TrimSpace(s); v != "" {
		return v
	}
	return fallback
}

// ParseBool parses boolean dari string value (case-insensitive)
func ParseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

// IsEmptyRow checks apakah semua cells di row kosong
func IsEmptyRow(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

// UniqueStrings removes duplicate strings dari slice
func UniqueStrings(items []string) []string {
	seen := make(map[string]bool, len(items))
	result := make([]string, 0, len(items))
	for _, s := range items {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
