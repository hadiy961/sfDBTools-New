package backup

import (
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
)

// DisplayFilterStats menampilkan statistik hasil pemfilteran database.
func (s *Service) DisplayFilterStats(stats *types.DatabaseFilterStats) {
	ui.DisplayFilterStats(stats, "backup", s.Log)
}

// DisplayBackupDBOptions menampilkan opsi backup dari field BackupDBOptions.
// untuk konfirmasi user sebelum melanjutkan proses backup.
func (s *Service) DisplayBackupDBOptions() (proceed bool, err error) {
	ui.PrintSubHeader("Opsi Backup")

	// Label untuk filename example tergantung mode
	filenameLabel := "Filename Example"
	if s.BackupDBOptions.Mode == "separated" || s.BackupDBOptions.Mode == "separate" {
		filenameLabel = "Filename Example (per DB)"
	}

	// Informasi Umum
	data := [][]string{
		{"Mode Backup", ui.ColorText(s.BackupDBOptions.Mode, ui.ColorCyan)},
		{"Output Directory", s.BackupDBOptions.OutputDir},
		{"Filename Pattern", helper.FixedBackupPattern},
		{filenameLabel, ui.ColorText(s.BackupDBOptions.File.Path, ui.ColorCyan)},
		{"Dry Run", fmt.Sprintf("%v", s.BackupDBOptions.DryRun)},
		{"Capture GTID", fmt.Sprintf("%v", s.BackupDBOptions.CaptureGTID)},
	}

	// Informasi Profile
	if s.BackupDBOptions.Profile.Name != "" {
		data = append(data, []string{"Profile", ui.ColorText(s.BackupDBOptions.Profile.Name, ui.ColorYellow)})
		data = append(data, []string{"HostName", s.BackupDBOptions.Profile.DBInfo.HostName})
		data = append(data, []string{"Host", fmt.Sprintf("%s:%d", s.BackupDBOptions.Profile.DBInfo.Host, s.BackupDBOptions.Profile.DBInfo.Port)})
		data = append(data, []string{"User", s.BackupDBOptions.Profile.DBInfo.User})
	}

	// Filter Options
	data = append(data, []string{"", ""}) // Empty row for separation
	data = append(data, []string{ui.ColorText("Filter Options", ui.ColorPurple), ""})
	data = append(data, []string{"Exclude System DB", fmt.Sprintf("%v", s.BackupDBOptions.Filter.ExcludeSystem)})
	data = append(data, []string{"Exclude User DB", fmt.Sprintf("%v", s.BackupDBOptions.Filter.ExcludeUser)})
	data = append(data, []string{"Exclude Empty DB", fmt.Sprintf("%v", s.BackupDBOptions.Filter.ExcludeEmpty)})
	data = append(data, []string{"Exclude Data DB", fmt.Sprintf("%v", s.BackupDBOptions.Filter.ExcludeData)})

	if len(s.BackupDBOptions.Filter.IncludeDatabases) > 0 {
		data = append(data, []string{"Include List", fmt.Sprintf("%d database", len(s.BackupDBOptions.Filter.IncludeDatabases))})
		if len(s.BackupDBOptions.Filter.IncludeDatabases) < 5 {
			for _, db := range s.BackupDBOptions.Filter.IncludeDatabases {
				data = append(data, []string{"  - " + db, ""})
			}
		}
	}

	if len(s.BackupDBOptions.Filter.ExcludeDatabases) > 0 {
		data = append(data, []string{"Exclude List", fmt.Sprintf("%d database", len(s.BackupDBOptions.Filter.ExcludeDatabases))})
		if len(s.BackupDBOptions.Filter.ExcludeDatabases) < 5 {
			for _, db := range s.BackupDBOptions.Filter.ExcludeDatabases {
				data = append(data, []string{"  - " + db, ""})
			}
		}
	}

	if s.BackupDBOptions.Filter.IncludeFile != "" {
		data = append(data, []string{"Include File", s.BackupDBOptions.Filter.IncludeFile})
	}

	if s.BackupDBOptions.Filter.ExcludeDBFile != "" {
		data = append(data, []string{"Exclude File", s.BackupDBOptions.Filter.ExcludeDBFile})
	}

	// Compression Options
	data = append(data, []string{"", ""}) // Empty row for separation
	data = append(data, []string{ui.ColorText("Compression", ui.ColorPurple), ""})
	data = append(data, []string{"Enabled", fmt.Sprintf("%v", s.BackupDBOptions.Compression.Enabled)})
	if s.BackupDBOptions.Compression.Enabled {
		data = append(data, []string{"Type", s.BackupDBOptions.Compression.Type})
		data = append(data, []string{"Level", fmt.Sprintf("%d", s.BackupDBOptions.Compression.Level)})
	}

	// Encryption Options
	data = append(data, []string{"", ""}) // Empty row for separation
	data = append(data, []string{ui.ColorText("Encryption", ui.ColorPurple), ""})
	data = append(data, []string{"Enabled", fmt.Sprintf("%v", s.BackupDBOptions.Encryption.Enabled)})
	if s.BackupDBOptions.Encryption.Enabled {
		data = append(data, []string{"Key Status", ui.ColorText("Configured", ui.ColorGreen)})
	}

	// Cleanup Options
	data = append(data, []string{"", ""}) // Empty row for separation
	data = append(data, []string{ui.ColorText("Cleanup", ui.ColorPurple), ""})
	data = append(data, []string{"Enabled", fmt.Sprintf("%v", s.BackupDBOptions.Cleanup.Enabled)})
	if s.BackupDBOptions.Cleanup.Enabled {
		data = append(data, []string{"Days to Keep", fmt.Sprintf("%d", s.BackupDBOptions.Cleanup.Days)})
		if s.BackupDBOptions.Cleanup.Pattern != "" {
			data = append(data, []string{"Pattern", s.BackupDBOptions.Cleanup.Pattern})
		}
		if s.BackupDBOptions.Cleanup.CleanupSchedule != "" {
			data = append(data, []string{"Schedule", s.BackupDBOptions.Cleanup.CleanupSchedule})
		}
		data = append(data, []string{"Background", fmt.Sprintf("%v", s.BackupDBOptions.Cleanup.Background)})
	}

	ui.FormatTable([]string{"Parameter", "Value"}, data)

	// Konfirmasi sebelum melanjutkan
	confirm, askErr := input.AskYesNo("Apakah Anda ingin melanjutkan?", true)
	if askErr != nil {
		s.Log.Error("gagal mendapatkan konfirmasi user: " + askErr.Error())
		return false, askErr
	}
	if !confirm {
		return false, types.ErrUserCancelled
	}
	s.Log.Info("Proses backup dilanjutkan.")
	return true, nil
}

func (s *Service) DisplayBackupResult(result *types.BackupResult) {
	ui.PrintSubHeader("Hasil Backup Database")

	// Summary statistik
	data := [][]string{
		{"Total Database Ditemukan", fmt.Sprintf("%d", result.TotalDatabases)},
		{"Total Database Dibackup", ui.ColorText(fmt.Sprintf("%d", result.SuccessfulBackups), ui.ColorGreen)},
		{"Total Gagal Dibackup", ui.ColorText(fmt.Sprintf("%d", result.FailedBackups), ui.ColorRed)},
		{"Total Waktu Proses", ui.ColorText(result.TotalTimeTaken.String(), ui.ColorCyan)},
	}
	ui.FormatTable([]string{"Kategori", "Jumlah"}, data)

	// Detail informasi backup sukses
	if len(result.BackupInfo) > 0 {
		ui.PrintSubHeader("Detail Backup yang Berhasil")

		for _, info := range result.BackupInfo {
			fmt.Println()

			// Status dengan warna
			statusText := info.Status
			statusColor := ui.ColorGreen
			if info.Status == "success_with_warnings" {
				statusText = "Sukses dengan Warning"
				statusColor = ui.ColorYellow
			}

			detailData := [][]string{
				{"Database", ui.ColorText(info.DatabaseName, ui.ColorCyan)},
				{"Status", ui.ColorText(statusText, statusColor)},
				{"File Output", info.OutputFile},
				{"Ukuran File", info.FileSizeHuman},
				{"Durasi Backup", info.Duration},
			}

			// Tambahkan informasi kompresi ratio jika ada
			if s.BackupDBOptions.Compression.Enabled {
				detailData = append(detailData, []string{"Kompresi", ui.ColorText("Enabled ("+s.BackupDBOptions.Compression.Type+")", ui.ColorGreen)})
			} else {
				detailData = append(detailData, []string{"Kompresi", ui.ColorText("Disabled", ui.ColorYellow)})
			}

			// Tambahkan informasi enkripsi
			if s.BackupDBOptions.Encryption.Enabled {
				detailData = append(detailData, []string{"Enkripsi", ui.ColorText("Enabled", ui.ColorGreen)})
			} else {
				detailData = append(detailData, []string{"Enkripsi", ui.ColorText("Disabled", ui.ColorYellow)})
			}

			ui.FormatTable([]string{"Parameter", "Nilai"}, detailData)

			// Tampilkan warning jika ada
			if info.Warnings != "" {
				ui.PrintWarning("\n⚠ Warning dari mysqldump:")
				fmt.Println(ui.ColorText(info.Warnings, ui.ColorYellow))
			}
		}
	}

	// Daftar database yang gagal
	if result.FailedBackups > 0 {
		ui.PrintSubHeader("Daftar Database Gagal Dibackup")

		if len(result.FailedDatabaseInfos) > 0 {
			for i, failedInfo := range result.FailedDatabaseInfos {
				fmt.Printf("%d. %s\n", i+1, ui.ColorText(failedInfo.DatabaseName, ui.ColorRed))
				fmt.Printf("   Error: %s\n", failedInfo.Error)
				fmt.Println()
			}
		} else if len(result.FailedDatabases) > 0 {
			for dbName, errMsg := range result.FailedDatabases {
				ui.PrintError("- " + dbName + ": " + errMsg)
			}
		}
	}
}
