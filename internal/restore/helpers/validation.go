// File : internal/restore/helpers/validation.go
// Deskripsi : Helper functions untuk validation
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package helpers

import (
	"context"
	"fmt"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
	"strings"
)

// ValidateNotPrimaryDatabase memvalidasi bahwa target database bukan database primary yang sudah ada
func ValidateNotPrimaryDatabase(ctx context.Context, client *database.Client, dbName string) error {
	dbLower := strings.ToLower(dbName)

	// Check prefix
	hasPrimaryPrefix := strings.HasPrefix(dbLower, "dbsf_nbc_") || strings.HasPrefix(dbLower, "dbsf_biznet_")
	if !hasPrimaryPrefix {
		return nil
	}

	// Check suffix
	hasNonPrimarySuffix := strings.Contains(dbLower, "_secondary") ||
		strings.HasSuffix(dbLower, "_dmart") ||
		strings.HasSuffix(dbLower, "_temp") ||
		strings.HasSuffix(dbLower, "_archive")

	if hasNonPrimarySuffix {
		return nil
	}

	// Check if database exists
	dbExists, err := client.CheckDatabaseExists(ctx, dbName)
	if err != nil {
		// Assume exists for safety
	} else if !dbExists {
		// Database doesn't exist yet, OK to create
		return nil
	}

	// Database exists and is primary - REJECT
	ui.PrintError("⛔ Restore ke database primary yang sudah ada tidak diizinkan!")
	ui.PrintError(fmt.Sprintf("   Database: %s", dbName))
	ui.PrintError("   Status: Database sudah ada di server")
	ui.PrintError("   Pattern: Database dengan awalan 'dbsf_nbc_' atau 'dbsf_biznet_'")
	ui.PrintError("   tanpa suffix '_secondary', '_dmart', '_temp', atau '_archive'")
	ui.PrintError("   dianggap sebagai database primary dan tidak boleh di-restore.")
	ui.PrintError("")
	ui.PrintInfo("💡 Saran:")
	ui.PrintInfo("   - Restore ke database secondary, temp, atau archive")
	ui.PrintInfo("   - Atau gunakan nama database yang berbeda")
	ui.PrintInfo("   - Database primary baru yang belum ada diperbolehkan untuk dibuat")

	return fmt.Errorf("restore ke database primary '%s' yang sudah ada tidak diizinkan", dbName)
}
