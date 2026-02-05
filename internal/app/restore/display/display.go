// File : internal/restore/display/display.go
// Deskripsi : Display functions untuk restore results
// Author : Hadiyatna Muflihun
// Tanggal : 17 Desember 2025
// Last Modified : 5 Januari 2026
package display

import (
	"fmt"
	"path/filepath"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/ui/print"
)

// ShowRestoreSingleResult menampilkan hasil restore single
func ShowRestoreSingleResult(result *restoremodel.RestoreResult) {
	print.PrintSubHeader("Hasil Restore")
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

	if result.SQLErrors > 0 || result.SQLWarnings > 0 {
		fmt.Printf("  %-20s: %d\n", "SQL Errors", result.SQLErrors)
		fmt.Printf("  %-20s: %d\n", "SQL Warnings", result.SQLWarnings)
	}

	fmt.Printf("  %-20s: %s\n", "Duration", result.Duration)
	status := "Berhasil"
	if !result.Success {
		status = "Gagal"
	} else if result.SQLErrors > 0 || result.SQLWarnings > 0 {
		status = "Berhasil (dengan peringatan)"
	}
	fmt.Printf("  %-20s: %s\n", "Status", status)
	fmt.Println()
}

// ShowRestorePrimaryResult menampilkan hasil restore primary
func ShowRestorePrimaryResult(result *restoremodel.RestoreResult) {
	print.PrintSubHeader("Hasil Restore Primary")
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

	if result.SQLErrors > 0 || result.SQLWarnings > 0 {
		fmt.Printf("  %-20s: %d\n", "SQL Errors", result.SQLErrors)
		fmt.Printf("  %-20s: %d\n", "SQL Warnings", result.SQLWarnings)
	}

	fmt.Printf("  %-20s: %s\n", "Duration", result.Duration)
	status := "Berhasil"
	if !result.Success {
		status = "Gagal"
	} else if result.SQLErrors > 0 || result.SQLWarnings > 0 {
		status = "Berhasil (dengan peringatan)"
	}
	fmt.Printf("  %-20s: %s\n", "Status", status)
	fmt.Println()
}

// ShowRestoreSecondaryResult menampilkan hasil restore secondary
func ShowRestoreSecondaryResult(result *restoremodel.RestoreResult) {
	print.PrintSubHeader("Hasil Restore Secondary")
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

	if result.DroppedDB {
		fmt.Printf("  %-20s: %s\n", "Database Dropped", "Ya")
	}
	if result.DroppedCompanion {
		fmt.Printf("  %-20s: %s\n", "Companion Dropped", "Ya")
	}

	if result.SQLErrors > 0 || result.SQLWarnings > 0 {
		fmt.Printf("  %-20s: %d\n", "SQL Errors", result.SQLErrors)
		fmt.Printf("  %-20s: %d\n", "SQL Warnings", result.SQLWarnings)
	}

	fmt.Printf("  %-20s: %s\n", "Duration", result.Duration)
	status := "Berhasil"
	if !result.Success {
		status = "Gagal"
	} else if result.SQLErrors > 0 || result.SQLWarnings > 0 {
		status = "Berhasil (dengan peringatan)"
	}
	fmt.Printf("  %-20s: %s\n", "Status", status)
	fmt.Println()
}

// ShowRestoreAllResult menampilkan hasil restore all databases
func ShowRestoreAllResult(result *restoremodel.RestoreResult) {
	print.PrintSubHeader("Hasil Restore All Databases")
	fmt.Println()

	fmt.Printf("  %-20s: %s\n", "Source File", result.SourceFile)

	if result.BackupFile != "" {
		fmt.Printf("  %-20s: %s\n", "Backup Pre-Restore", result.BackupFile)
		fmt.Printf("  %-20s: %s\n", "Backup Directory", filepath.Dir(result.BackupFile))
	}

	fmt.Printf("  %-20s: %s\n", "Duration", result.Duration)

	if result.SQLErrors > 0 || result.SQLWarnings > 0 {
		fmt.Printf("  %-20s: %d\n", "SQL Errors", result.SQLErrors)
		fmt.Printf("  %-20s: %d\n", "SQL Warnings", result.SQLWarnings)
	}

	if result.Success {
		status := "Berhasil"
		if result.SQLErrors > 0 || result.SQLWarnings > 0 {
			status = "Berhasil (dengan peringatan)"
		}
		fmt.Printf("  %-20s: %s\n", "Status", status)
	} else {
		fmt.Printf("  %-20s: %s\n", "Status", "Gagal")
	}

	fmt.Println()
}

// ShowRestoreCustomResult menampilkan hasil restore custom (DB + DMART)
func ShowRestoreCustomResult(result *restoremodel.RestoreResult) {
	print.PrintSubHeader("Hasil Restore Custom")
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

	if result.SQLErrors > 0 || result.SQLWarnings > 0 {
		fmt.Printf("  %-20s: %d\n", "SQL Errors", result.SQLErrors)
		fmt.Printf("  %-20s: %d\n", "SQL Warnings", result.SQLWarnings)
	}

	fmt.Printf("  %-20s: %s\n", "Duration", result.Duration)
	status := "Berhasil"
	if !result.Success {
		status = "Gagal"
	} else if result.SQLErrors > 0 || result.SQLWarnings > 0 {
		status = "Berhasil (dengan peringatan)"
	}
	fmt.Printf("  %-20s: %s\n", "Status", status)
	fmt.Println()
}
