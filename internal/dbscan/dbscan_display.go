package dbscan

import (
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
)

// DisplayScanOptions menampilkan opsi scanning yang sedang aktif dan meminta konfirmasi.
// Mengembalikan:
// - proceed=true jika pengguna memilih untuk melanjutkan
// - proceed=false jika pengguna membatalkan (tanpa error)
// - err != nil jika terjadi kegagalan saat meminta input
func (s *Service) DisplayScanOptions() (proceed bool, err error) {
	ui.PrintSubHeader("Opsi Scanning")
	targetConn := s.getTargetDBConfig()

	data := [][]string{
		{"Exclude System DB", fmt.Sprintf("%v", s.ScanOptions.ExcludeSystem)},
		{"Include List", fmt.Sprintf("%d database", len(s.ScanOptions.IncludeList))},
		{"Exclude List", fmt.Sprintf("%d database", len(s.ScanOptions.ExcludeList))},
		{"Save to DB", fmt.Sprintf("%v", s.ScanOptions.SaveToDB)},
		{"Display Results", fmt.Sprintf("%v", s.ScanOptions.DisplayResults)},
	}

	if s.ScanOptions.SaveToDB {
		targetInfo := fmt.Sprintf("%s@%s:%d/%s",
			targetConn.User, targetConn.Host, targetConn.Port, targetConn.Database)
		data = append(data, []string{"Target DB", targetInfo})
	}

	if s.ScanOptions.Mode == "single" && s.ScanOptions.SourceDatabase != "" {
		data = append(data, []string{"Source Database", s.ScanOptions.SourceDatabase})
	}

	ui.FormatTable([]string{"Parameter", "Value"}, data)

	// Konfirmasi sebelum melanjutkan
	confirm, askErr := input.AskYesNo("Apakah Anda ingin melanjutkan?", true)
	if askErr != nil {
		s.Logger.Error("gagal mendapatkan konfirmasi user: " + askErr.Error())
		return false, askErr
	}
	if !confirm {
		return false, types.ErrUserCancelled
	}
	s.Logger.Info("Proses scanning dilanjutkan.")
	return true, nil
}

// DisplayFilterStats menampilkan statistik hasil pemfilteran database.
func (s *Service) DisplayFilterStats(stats *types.DatabaseFilterStats) {
	ui.DisplayFilterStats(stats, "scan", s.Logger)
}

// DisplayDetailResults menampilkan detail hasil scanning
func (s *Service) DisplayDetailResults(detailsMap map[string]types.DatabaseDetailInfo) {
	ui.PrintHeader("DETAIL HASIL SCANNING")

	headers := []string{"Database", "Size", "Tables", "Procedures", "Functions", "Views", "Grants", "Status"}
	var rows [][]string

	for _, detail := range detailsMap {
		status := ui.ColorText("✓ OK", ui.ColorGreen)
		if detail.Error != "" {
			status = ui.ColorText("✗ Error", ui.ColorRed)
		}

		rows = append(rows, []string{
			detail.DatabaseName,
			detail.SizeHuman,
			fmt.Sprintf("%d", detail.TableCount),
			fmt.Sprintf("%d", detail.ProcedureCount),
			fmt.Sprintf("%d", detail.FunctionCount),
			fmt.Sprintf("%d", detail.ViewCount),
			fmt.Sprintf("%d", detail.UserGrantCount),
			status,
		})
	}

	ui.FormatTable(headers, rows)
}

// DisplayScanResult menampilkan hasil scanning
func (s *Service) DisplayScanResult(result *types.ScanResult) {
	ui.PrintHeader("HASIL SCANNING")

	data := [][]string{
		{"Total Database", fmt.Sprintf("%d", result.TotalDatabases)},
		{"Berhasil", ui.ColorText(fmt.Sprintf("%d", result.SuccessCount), ui.ColorGreen)},
		{"Gagal", ui.ColorText(fmt.Sprintf("%d", result.FailedCount), ui.ColorRed)},
		{"Durasi", result.Duration},
	}

	headers := []string{"Metrik", "Nilai"}
	ui.FormatTable(headers, data)

	if len(result.Errors) > 0 {
		ui.PrintWarning(fmt.Sprintf("Terdapat %d error saat menyimpan ke database:", len(result.Errors)))
		for _, errMsg := range result.Errors {
			fmt.Printf("  • %s\n", errMsg)
		}
	}
}

// LogDetailResults menulis detail hasil scanning ke logger (untuk background mode)
func (s *Service) LogDetailResults(detailsMap map[string]types.DatabaseDetailInfo) {
	s.Logger.Info("=== Detail Hasil Scanning ===")

	for dbName, detail := range detailsMap {
		if detail.Error != "" {
			s.Logger.Warnf("Database: %s - Status: ERROR - %s", dbName, detail.Error)
		}
	}
}
