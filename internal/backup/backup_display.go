package backup

import (
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/compress"
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

	// Dapatkan hostname dari database server (bukan hostname lokal)
	dbHostname := s.BackupDBOptions.Profile.DBInfo.Host

	// Convert compression type dari string ke compress.CompressionType
	// Jika compression disabled, gunakan CompressionNone
	compressionType := compress.CompressionNone
	if s.BackupDBOptions.Compression.Enabled {
		compressionType = compress.CompressionType(s.BackupDBOptions.Compression.Type)
	}

	// Generate filename berdasarkan pattern dari config
	filename, err := helper.GenerateBackupFilename(
		s.Config.Backup.Output.NamePattern,
		"", // database kosong untuk preview, akan diisi saat backup actual
		s.BackupDBOptions.Mode,
		dbHostname,
		compressionType,
		s.BackupDBOptions.Encryption.Enabled,
	)
	if err != nil {
		s.Log.Warn("gagal generate filename preview: " + err.Error())
		filename = "error_generating_filename"
	}

	// Generate output directory berdasarkan config
	outputDir, err := helper.GenerateBackupDirectory(
		s.Config.Backup.Output.BaseDirectory,
		s.Config.Backup.Output.Structure.Pattern,
		dbHostname,
	)
	if err != nil {
		s.Log.Warn("gagal generate output directory: " + err.Error())
		outputDir = s.Config.Backup.Output.BaseDirectory
	}

	// Simpan ke options untuk digunakan nanti
	s.BackupDBOptions.OutputDir = outputDir
	s.BackupDBOptions.NamePattern = s.Config.Backup.Output.NamePattern

	// Informasi Umum
	data := [][]string{
		{"Mode Backup", ui.ColorText(s.BackupDBOptions.Mode, ui.ColorCyan)},
		{"Output Directory", outputDir},
		{"Filename Pattern", s.Config.Backup.Output.NamePattern},
		{"Filename Example", ui.ColorText(filename, ui.ColorCyan)},
		{"Background Mode", fmt.Sprintf("%v", s.BackupDBOptions.Background)},
		{"Dry Run", fmt.Sprintf("%v", s.BackupDBOptions.DryRun)},
		{"Capture GTID", fmt.Sprintf("%v", s.BackupDBOptions.CaptureGTID)},
	}

	// Informasi Profile
	if s.BackupDBOptions.Profile.Name != "" {
		data = append(data, []string{"Profile", ui.ColorText(s.BackupDBOptions.Profile.Name, ui.ColorYellow)})
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
	data := [][]string{
		{"Total Database Ditemukan", fmt.Sprintf("%d", result.TotalDatabases)},
		{"Total Database Dibackup", fmt.Sprintf("%d", result.SuccessfulBackups)},
		{"Total Gagal Dibackup", fmt.Sprintf("%d", result.FailedBackups)},
	}
	ui.FormatTable([]string{"Kategori", "Jumlah"}, data)

	if result.FailedBackups > 0 {
		ui.PrintSubHeader("Daftar Database Gagal Dibackup")
		for _, dbName := range result.FailedDatabases {
			ui.PrintError("- " + dbName)
		}
	}
}
