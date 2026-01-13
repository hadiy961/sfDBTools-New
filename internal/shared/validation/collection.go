// File : internal/shared/validation/collection.go
// Deskripsi : Helper validasi untuk koleksi (slice) sederhana
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package validation

import "strings"

// UniqueStrings mengembalikan list string unik (preserve order).
//
// Example:
//
//	items := []string{"a", "b", "a", "c", "b"}
//	unique := UniqueStrings(items) // ["a", "b", "c"]
func UniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

// SubtractStrings mengembalikan items dari list1 yang tidak ada di list2.
//
// Example:
//
//	list1 := []string{"a", "b", "c"}
//	list2 := []string{"b"}
//	result := SubtractStrings(list1, list2) // ["a", "c"]
func SubtractStrings(list1, list2 []string) []string {
	exclude := make(map[string]struct{}, len(list2))
	for _, item := range list2 {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		exclude[item] = struct{}{}
	}

	result := make([]string, 0, len(list1))
	for _, item := range list1 {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, excluded := exclude[item]; excluded {
			continue
		}
		result = append(result, item)
	}
	return result
}
