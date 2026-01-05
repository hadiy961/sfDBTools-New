// File : internal/restore/helpers/validation.go
// Deskripsi : Helper functions untuk validation
// Author : Hadiyatna Muflihun
// Tanggal : 17 Desember 2025
// Last Modified : 17 Desember 2025

package helpers

import (
	"context"
	"fmt"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
	"strings"
)

func IsPrimaryDatabaseName(dbName string) bool {
	dbLower := strings.ToLower(strings.TrimSpace(dbName))
	if dbLower == "" {
		return false
	}

	hasPrimaryPrefix := strings.HasPrefix(dbLower, consts.PrimaryPrefixNBC) || strings.HasPrefix(dbLower, consts.PrimaryPrefixBiznet)
	if !hasPrimaryPrefix {
		return false
	}

	// Primary harus TANPA suffix non-primary.
	if strings.Contains(dbLower, consts.SecondarySuffix) ||
		strings.HasSuffix(dbLower, consts.SuffixDmart) ||
		strings.HasSuffix(dbLower, consts.SuffixTemp) ||
		strings.HasSuffix(dbLower, consts.SuffixArchive) {
		return false
	}

	// Harus ada client-code setelah prefix.
	if strings.HasPrefix(dbLower, consts.PrimaryPrefixNBC) {
		return len(dbLower) > len(consts.PrimaryPrefixNBC)
	}
	return len(dbLower) > len(consts.PrimaryPrefixBiznet)
}

// ValidatePrimaryDatabaseName memvalidasi bahwa nama database adalah primary:
// dbsf_nbc_{client-code} atau dbsf_biznet_{client-code} (tanpa suffix non-primary).
func ValidatePrimaryDatabaseName(dbName string) error {
	if IsPrimaryDatabaseName(dbName) {
		return nil
	}

	ui.PrintError("⛔ Restore primary hanya diizinkan ke database primary!")
	ui.PrintError(fmt.Sprintf("   Database: %s", dbName))
	ui.PrintError("   Pattern: dbsf_nbc_{client-code} atau dbsf_biznet_{client-code}")
	ui.PrintError("   Catatan: tanpa suffix '_secondary', '_dmart', '_temp', atau '_archive'")
	return fmt.Errorf("target database '%s' bukan database primary yang valid", dbName)
}

// ValidateNotPrimaryDatabaseName memvalidasi bahwa nama database BUKAN primary.
// Digunakan untuk mode restore single: tidak boleh sama sekali ke database primary.
func ValidateNotPrimaryDatabaseName(dbName string) error {
	if !IsPrimaryDatabaseName(dbName) {
		return nil
	}

	ui.PrintError("⛔ Restore single ke database primary tidak diizinkan!")
	ui.PrintError(fmt.Sprintf("   Database: %s", dbName))
	ui.PrintError("   Pattern primary: dbsf_nbc_{client-code} atau dbsf_biznet_{client-code}")
	ui.PrintInfo("💡 Saran: gunakan mode 'restore primary' atau restore ke database non-primary (mis: _secondary/_temp/_archive)")
	return fmt.Errorf("restore single ke database primary '%s' tidak diizinkan", dbName)
}

// ValidateNotPrimaryDatabase memvalidasi bahwa target database bukan database primary.
// NOTE: Sesuai aturan safety terbaru, restore single tidak boleh sama sekali ke database primary.
func ValidateNotPrimaryDatabase(ctx context.Context, client *database.Client, dbName string) error {
	_ = ctx
	_ = client
	return ValidateNotPrimaryDatabaseName(dbName)
}
