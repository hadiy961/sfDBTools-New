// File : internal/restore/restore_display.go
// Deskripsi : Display functions untuk restore options dan results
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-10
// Last Modified : 2025-11-10

package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

// DisplayRestoreOptions menampilkan opsi restore dan meminta konfirmasi.
// Mengembalikan:
// - proceed=true jika pengguna memilih untuk melanjutkan
// - proceed=false jika pengguna membatalkan (tanpa error)
// - err != nil jika terjadi kegagalan saat meminta input
func (s *Service) DisplayRestoreOptions() (proceed bool, err error) {
	ui.PrintSubHeader("Opsi Restore")

	// Get file info untuk display
	fileInfo := "unknown"
	sourceFile := s.RestoreOptions.SourceFile
	if stat, err := os.Stat(sourceFile); err == nil {
		fileSize := global.FormatFileSize(stat.Size())
		fileInfo = fmt.Sprintf("%s (%s)", filepath.Base(sourceFile), fileSize)
	}

	// Build data rows untuk table
	data := [][]string{
		{"Source File", ui.ColorText(fileInfo, ui.ColorCyan)},
		{"Dry Run", fmt.Sprintf("%v", s.RestoreOptions.DryRun)},
		{"Skip Backup", fmt.Sprintf("%v", s.RestoreOptions.SkipBackup)},
		{"Drop Target", fmt.Sprintf("%v", s.RestoreOptions.DropTarget)},
	}

	// Add target profile info jika tersedia
	if s.TargetProfile != nil {
		data = append(data, []string{"", ""}) // Empty row for separation
		data = append(data, []string{ui.ColorText("Target Database", ui.ColorPurple), ""})
		data = append(data, []string{"Profile", ui.ColorText(s.TargetProfile.Name, ui.ColorYellow)})
		data = append(data, []string{"Host", fmt.Sprintf("%s:%d", s.TargetProfile.DBInfo.Host, s.TargetProfile.DBInfo.Port)})
		data = append(data, []string{"User", s.TargetProfile.DBInfo.User})

		// Add restore mode specific info
		if s.RestoreEntry != nil && s.RestoreEntry.RestoreMode == "single" {
			targetDB := s.RestoreOptions.TargetDB
			if targetDB == "" {
				targetDB = "(extracted from filename)"
			}
			data = append(data, []string{"Target Database", ui.ColorText(targetDB, ui.ColorCyan)})
		}
	}

	// Display encryption info jika ada
	if s.RestoreOptions.EncryptionKey != "" {
		data = append(data, []string{"", ""}) // Empty row for separation
		data = append(data, []string{ui.ColorText("Encryption", ui.ColorPurple), ""})
		data = append(data, []string{"Encryption Key", ui.ColorText("*** (configured)", ui.ColorGreen)})
	}

	// Print table
	ui.FormatTable([]string{"Option", "Value"}, data)
	fmt.Println()

	// Ask for confirmation
	if s.RestoreOptions.Force {
		// Force mode - no confirmation needed
		s.Log.Info("✓ Running in force mode (no confirmation needed)")
		fmt.Println()
		return true, nil
	}

	// Prompt for confirmation using AskYesNo
	confirmed, err := input.AskYesNo("Lanjutkan restore dengan konfigurasi di atas?", true)
	if err != nil {
		return false, validation.HandleInputError(err)
	}

	if !confirmed {
		return false, types.ErrUserCancelled
	}

	fmt.Println()
	return true, nil
}

// DisplayRestoreResult menampilkan hasil restore
func (s *Service) DisplayRestoreResult(result types.RestoreResult) {
	ui.PrintSubHeader("Hasil Restore")

	data := [][]string{
		{"Total Databases", fmt.Sprintf("%d", result.TotalDatabases)},
		{"Successful", ui.ColorText(fmt.Sprintf("%d", result.SuccessfulRestore), ui.ColorGreen)},
		{"Failed", ui.ColorText(fmt.Sprintf("%d", result.FailedRestore), ui.ColorRed)},
		{"Total Time", result.TotalTimeTaken.String()},
	}

	if result.PreBackupFile != "" {
		data = append(data, []string{"", ""})
		data = append(data, []string{"Pre-Backup File", filepath.Base(result.PreBackupFile)})
	}

	ui.FormatTable([]string{"Metric", "Value"}, data)
	fmt.Println()

	// Display detail per database jika ada
	if len(result.RestoreInfo) > 0 {
		ui.PrintSubHeader("Detail Restore Per Database")
		detailData := [][]string{}

		for _, info := range result.RestoreInfo {
			status := info.Status
			if status == "success" {
				// Jika ada warning, tampilkan dengan highlight
				if info.Warnings != "" {
					status = ui.ColorText("✓ "+status+" (with warnings)", ui.ColorYellow)
				} else {
					status = ui.ColorText("✓ "+status, ui.ColorGreen)
				}
			} else if status == "failed" {
				status = ui.ColorText("✗ "+status, ui.ColorRed)
			}

			detailData = append(detailData, []string{
				info.DatabaseName,
				info.TargetDatabase,
				status,
				info.Duration,
			})
		}

		ui.FormatTable([]string{"Source DB", "Target DB", "Status", "Duration"}, detailData)
		fmt.Println()

		// Display warnings detail jika ada
		hasWarnings := false
		for _, info := range result.RestoreInfo {
			if info.Warnings != "" {
				if !hasWarnings {
					ui.PrintSubHeader("Warnings")
					hasWarnings = true
				}
				fmt.Printf("• %s:\n", info.DatabaseName)
				fmt.Printf("  %s\n", ui.ColorText(info.Warnings, ui.ColorYellow))
			}
		}
		if hasWarnings {
			fmt.Println()
		}
	}

	// Display errors jika ada
	if len(result.Errors) > 0 {
		ui.PrintSubHeader("Errors & Warnings")
		for i, errMsg := range result.Errors {
			isWarning := len(errMsg) > 7 && errMsg[:8] == "WARNING:"
			if isWarning {
				fmt.Printf("%d. %s\n", i+1, ui.ColorText(errMsg, ui.ColorYellow))
			} else {
				fmt.Printf("%d. %s\n", i+1, ui.ColorText(errMsg, ui.ColorRed))
			}
		}
		fmt.Println()
	}
}
