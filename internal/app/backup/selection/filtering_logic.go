// File : internal/app/backup/selection/filtering_logic.go
// Deskripsi : Centralized filtering logic dengan interface pattern untuk extensibility
// Author : Hadiyatna Muflihun
// Tanggal : 20 Januari 2026
// Last Modified : 20 Januari 2026

package selection

import (
	"strings"

	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
)

// ============================================================================
// Filter Interface & Implementations (ISP-compliant)
// ============================================================================

// DatabaseFilter adalah interface untuk memfilter database berdasarkan kriteria tertentu.
// Interface ini memungkinkan extensibility dan mudah di-test.
type DatabaseFilter interface {
	Match(dbName string) bool
}

// ============================================================================
// Basic Filters
// ============================================================================

// SystemDBFilter memfilter database system (mysql, information_schema, dll).
type SystemDBFilter struct{}

func (f SystemDBFilter) Match(dbName string) bool {
	dbLower := strings.ToLower(dbName)
	_, isSystem := domain.SystemDatabases[dbLower]
	return !isSystem // Return true untuk non-system DB
}

// SuffixFilter memfilter database berdasarkan suffix tertentu (exclude).
type SuffixFilter struct {
	ExcludedSuffixes []string
}

func (f SuffixFilter) Match(dbName string) bool {
	dbLower := strings.ToLower(dbName)
	for _, suffix := range f.ExcludedSuffixes {
		if strings.HasSuffix(dbLower, suffix) {
			return false
		}
	}
	return true
}

// ClientCodeFilter memfilter database berdasarkan client code untuk mode primary.
// Pattern: dbsf_nbc_{client_code}
type ClientCodeFilter struct {
	ClientCode string
}

func (f ClientCodeFilter) Match(dbName string) bool {
	if f.ClientCode == "" {
		return true
	}

	dbLower := strings.ToLower(dbName)
	targetPattern := consts.PrimaryPrefixNBC + strings.ToLower(f.ClientCode)

	// Match exact atau dengan underscore separator
	if dbLower == targetPattern || strings.HasPrefix(dbLower, targetPattern+"_") {
		// Exclude secondary, dmart, temp, archive
		if !strings.Contains(dbLower, consts.SecondarySuffix) &&
			!strings.HasSuffix(dbLower, consts.SuffixDmart) &&
			!strings.HasSuffix(dbLower, consts.SuffixTemp) &&
			!strings.HasSuffix(dbLower, consts.SuffixArchive) {
			return true
		}
	}
	return false
}

// PrimaryModeFilter memfilter database untuk mode primary.
type PrimaryModeFilter struct{}

func (f PrimaryModeFilter) Match(dbName string) bool {
	dbLower := strings.ToLower(dbName)

	// Exclude secondary, dmart, temp, archive
	if strings.Contains(dbLower, consts.SecondarySuffix) ||
		strings.HasSuffix(dbLower, consts.SuffixDmart) ||
		strings.HasSuffix(dbLower, consts.SuffixTemp) ||
		strings.HasSuffix(dbLower, consts.SuffixArchive) {
		return false
	}

	// Only NBC databases
	return strings.HasPrefix(dbLower, consts.PrimaryPrefixNBC)
}

// SecondaryModeFilter memfilter database untuk mode secondary.
type SecondaryModeFilter struct{}

func (f SecondaryModeFilter) Match(dbName string) bool {
	dbLower := strings.ToLower(dbName)

	// Exclude dmart, temp, archive
	if strings.HasSuffix(dbLower, consts.SuffixDmart) ||
		strings.HasSuffix(dbLower, consts.SuffixTemp) ||
		strings.HasSuffix(dbLower, consts.SuffixArchive) {
		return false
	}

	// Must be NBC database AND contain secondary suffix
	return strings.HasPrefix(dbLower, consts.PrimaryPrefixNBC) &&
		strings.Contains(dbLower, consts.SecondarySuffix)
}

// SecondaryClientCodeAndInstanceFilter memfilter database secondary berdasarkan client code dan instance.
// Pattern: dbsf_nbc_{client_code}_secondary_{instance}
type SecondaryClientCodeAndInstanceFilter struct {
	ClientCode string
	Instance   string
}

func (f SecondaryClientCodeAndInstanceFilter) Match(dbName string) bool {
	dbLower := strings.ToLower(dbName)

	// Exclude dmart, temp, archive
	if strings.HasSuffix(dbLower, consts.SuffixDmart) ||
		strings.HasSuffix(dbLower, consts.SuffixTemp) ||
		strings.HasSuffix(dbLower, consts.SuffixArchive) {
		return false
	}

	// Case 1: Only instance specified
	if f.ClientCode == "" && f.Instance != "" {
		targetSuffix := consts.SecondarySuffix + "_" + strings.ToLower(f.Instance)
		return strings.Contains(dbLower, targetSuffix)
	}

	// Case 2: Only client code specified
	if f.ClientCode != "" && f.Instance == "" {
		targetPattern := consts.PrimaryPrefixNBC + strings.ToLower(f.ClientCode) + consts.SecondarySuffix
		return strings.Contains(dbLower, targetPattern)
	}

	// Case 3: Both specified
	if f.ClientCode != "" && f.Instance != "" {
		targetPattern := consts.PrimaryPrefixNBC + strings.ToLower(f.ClientCode) + consts.SecondarySuffix + "_" + strings.ToLower(f.Instance)
		return dbLower == targetPattern || strings.HasPrefix(dbLower, targetPattern+"_")
	}

	// Case 4: Neither specified - match all
	return true
}

// ============================================================================
// Composite Filter
// ============================================================================

// AndFilter menggabungkan multiple filters dengan logika AND.
// Database harus match semua filter untuk pass.
type AndFilter struct {
	Filters []DatabaseFilter
}

func (f AndFilter) Match(dbName string) bool {
	for _, filter := range f.Filters {
		if !filter.Match(dbName) {
			return false
		}
	}
	return true
}

// ============================================================================
// Filter Application Functions (untuk backward compatibility)
// ============================================================================

// FilterCandidatesByMode memfilter database candidates berdasarkan backup mode.
func FilterCandidatesByMode(dbFiltered []string, mode string) []string {
	var filters []DatabaseFilter

	// Base filters untuk semua mode
	filters = append(filters, SystemDBFilter{})
	filters = append(filters, SuffixFilter{
		ExcludedSuffixes: []string{consts.SuffixTemp, consts.SuffixArchive},
	})

	// Mode-specific filters
	switch mode {
	case consts.ModePrimary:
		filters = append(filters, PrimaryModeFilter{})
	case consts.ModeSecondary:
		filters = append(filters, SecondaryModeFilter{})
	case consts.ModeSingle:
		// Single mode: exclude dmart only
		filters = append(filters, SuffixFilter{
			ExcludedSuffixes: []string{consts.SuffixDmart},
		})
	}

	return ApplyFilters(dbFiltered, AndFilter{Filters: filters})
}

// FilterCandidatesByClientCode memfilter database berdasarkan client code untuk mode primary.
// Pattern: dbsf_nbc_{client_code}
func FilterCandidatesByClientCode(databases []string, clientCode string) []string {
	if clientCode == "" {
		return databases
	}

	return ApplyFilters(databases, ClientCodeFilter{ClientCode: clientCode})
}

// FilterSecondaryByClientCodeAndInstance memfilter database secondary berdasarkan client code dan instance.
// Pattern: dbsf_nbc_{client_code}_secondary_{instance}
func FilterSecondaryByClientCodeAndInstance(databases []string, clientCode, instance string) []string {
	return ApplyFilters(databases, SecondaryClientCodeAndInstanceFilter{
		ClientCode: clientCode,
		Instance:   instance,
	})
}

// ApplyFilters applies a filter to a list of databases and returns filtered result.
func ApplyFilters(databases []string, filter DatabaseFilter) []string {
	filtered := make([]string, 0, len(databases))
	for _, db := range databases {
		if filter.Match(db) {
			filtered = append(filtered, db)
		}
	}
	return filtered
}
