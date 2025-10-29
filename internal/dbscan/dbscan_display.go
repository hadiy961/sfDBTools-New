package dbscan

import (
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
)

// DisplayScanOptions menampilkan opsi scanning yang sedang aktif.
func (s *Service) DisplayScanOptions() {
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
}

// DisplayFilterStats menampilkan statistik hasil pemfilteran database.
func (s *Service) DisplayFilterStats(stats *types.DatabaseFilterStats) {
	ui.PrintSubHeader("Statistik Filtering Database")
	data := [][]string{
		{"Total Ditemukan", fmt.Sprintf("%d", stats.TotalFound)},
		{"Akan di-scan", ui.ColorText(fmt.Sprintf("%d", stats.ToScan), ui.ColorGreen)},
		{"Dikecualikan (Sistem)", fmt.Sprintf("%d", stats.ExcludedSystem)},
		{"Dikecualikan (Exclude List)", fmt.Sprintf("%d", stats.ExcludedByList)},
		{"Dikecualikan (Bukan di Include List)", fmt.Sprintf("%d", stats.ExcludedByFile)},
		{"Dikecualikan (Nama Kosong)", fmt.Sprintf("%d", stats.ExcludedEmpty)},
	}
	ui.FormatTable([]string{"Kategori", "Jumlah"}, data)
}

// DisplayDetailResults menampilkan detail hasil scanning
func (s *Service) DisplayDetailResults(detailsMap map[string]database.DatabaseDetailInfo) {
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
func (s *Service) LogDetailResults(detailsMap map[string]database.DatabaseDetailInfo) {
	s.Logger.Info("=== Detail Hasil Scanning ===")

	for dbName, detail := range detailsMap {
		if detail.Error != "" {
			s.Logger.Warnf("Database: %s - Status: ERROR - %s", dbName, detail.Error)
		}
	}
}
