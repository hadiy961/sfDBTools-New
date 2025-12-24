// File : internal/restore/display/display.go
// Deskripsi : Display functions untuk restore results
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package display

import (
	"fmt"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/ui"
)

// ShowRestoreSingleResult menampilkan hasil restore single
func ShowRestoreSingleResult(result *types.RestoreResult) {
	ui.PrintSubHeader("Hasil Restore")
	fmt.Println()

	fmt.Printf("  %-20s: %s\n", "Target Database", result.TargetDB)
	fmt.Printf("  %-20s: %s\n", "Source File", result.SourceFile)

	if result.BackupFile != "" {
		fmt.Printf("  %-20s: %s\n", "Backup Pre-Restore", result.BackupFile)
		fmt.Printf("  %-20s: %s\n", "Backup Directory", filepath.Dir(result.BackupFile))
	}

	if result.DroppedDB {
		fmt.Printf("  %-20s: %s\n", "Database Dropped", "Ya")
	}

	if result.GrantsFile != "" {
		grantsStatus := "Ya"
		if !result.GrantsRestored {
			grantsStatus = "Gagal"
		}
		fmt.Printf("  %-20s: %s (%s)\n", "User Grants", filepath.Base(result.GrantsFile), grantsStatus)
	}

	fmt.Printf("  %-20s: %s\n", "Duration", result.Duration)
	fmt.Printf("  %-20s: %s\n", "Status", "Berhasil")
	fmt.Println()
}

// ShowRestorePrimaryResult menampilkan hasil restore primary
func ShowRestorePrimaryResult(result *types.RestoreResult) {
	ui.PrintSubHeader("Hasil Restore Primary")
	fmt.Println()

	fmt.Printf("  %-20s: %s\n", "Target Database", result.TargetDB)
	fmt.Printf("  %-20s: %s\n", "Source File", result.SourceFile)

	if result.CompanionDB != "" && result.CompanionFile != "" {
		fmt.Printf("  %-20s: %s\n", "Companion Database", result.CompanionDB)
		fmt.Printf("  %-20s: %s\n", "Companion File", result.CompanionFile)
	}

	if result.BackupFile != "" {
		fmt.Printf("  %-20s: %s\n", "Backup Pre-Restore", result.BackupFile)
		fmt.Printf("  %-20s: %s\n", "Backup Directory", filepath.Dir(result.BackupFile))
	}

	if result.CompanionBackup != "" {
		fmt.Printf("  %-20s: %s\n", "Companion Backup", filepath.Base(result.CompanionBackup))
	}

	if result.DroppedDB {
		fmt.Printf("  %-20s: %s\n", "Database Dropped", "Ya")
	}

	if result.DroppedCompanion {
		fmt.Printf("  %-20s: %s\n", "Companion Dropped", "Ya")
	}

	if result.GrantsFile != "" {
		grantsStatus := "Ya"
		if !result.GrantsRestored {
			grantsStatus = "Gagal"
		}
		fmt.Printf("  %-20s: %s (%s)\n", "User Grants", filepath.Base(result.GrantsFile), grantsStatus)
	}

	fmt.Printf("  %-20s: %s\n", "Duration", result.Duration)
	fmt.Printf("  %-20s: %s\n", "Status", "Berhasil")
	fmt.Println()
}

// ShowRestoreAllResult menampilkan hasil restore all databases
func ShowRestoreAllResult(result *types.RestoreResult) {
	ui.PrintSubHeader("Hasil Restore All Databases")
	fmt.Println()

	fmt.Printf("  %-20s: %s\n", "Source File", result.SourceFile)

	if result.BackupFile != "" {
		fmt.Printf("  %-20s: %s\n", "Backup Pre-Restore", result.BackupFile)
		fmt.Printf("  %-20s: %s\n", "Backup Directory", filepath.Dir(result.BackupFile))
	}

	fmt.Printf("  %-20s: %s\n", "Duration", result.Duration)

	if result.Success {
		fmt.Printf("  %-20s: %s\n", "Status", "Berhasil")
	} else {
		fmt.Printf("  %-20s: %s\n", "Status", "Gagal")
	}

	fmt.Println()
}

// ShowRestoreCustomResult menampilkan hasil restore custom (DB + DMART)
func ShowRestoreCustomResult(result *types.RestoreResult) {
	ui.PrintSubHeader("Hasil Restore Custom")
	fmt.Println()

	fmt.Printf("  %-20s: %s\n", "Target Database", result.TargetDB)
	fmt.Printf("  %-20s: %s\n", "Source File", result.SourceFile)

	if result.CompanionDB != "" && result.CompanionFile != "" {
		fmt.Printf("  %-20s: %s\n", "Database DMART", result.CompanionDB)
		fmt.Printf("  %-20s: %s\n", "DMART File", result.CompanionFile)
	}

	if result.BackupFile != "" {
		fmt.Printf("  %-20s: %s\n", "Backup Pre-Restore", result.BackupFile)
		fmt.Printf("  %-20s: %s\n", "Backup Directory", filepath.Dir(result.BackupFile))
	}
	if result.CompanionBackup != "" {
		fmt.Printf("  %-20s: %s\n", "Backup DMART", filepath.Base(result.CompanionBackup))
	}

	if result.DroppedDB {
		fmt.Printf("  %-20s: %s\n", "Database Dropped", "Ya")
	}
	if result.DroppedCompanion {
		fmt.Printf("  %-20s: %s\n", "DMART Dropped", "Ya")
	}

	fmt.Printf("  %-20s: %s\n", "Duration", result.Duration)
	status := "Berhasil"
	if !result.Success {
		status = "Gagal"
	}
	fmt.Printf("  %-20s: %s\n", "Status", status)
	fmt.Println()
}
