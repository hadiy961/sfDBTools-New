// File : internal/app/dbscan/display.go
// Deskripsi : Display functions untuk database scanning results
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 5 Januari 2026

package dbscan

import (
	"fmt"
	"sfDBTools/internal/app/dbscan/helpers"
	dbscanmodel "sfDBTools/internal/app/dbscan/model"
	"sfDBTools/internal/domain"
	"sfDBTools/internal/ui/print"
	"sfDBTools/internal/ui/prompt"
	"sfDBTools/internal/ui/table"
	"sfDBTools/pkg/validation"
)

// DisplayScanOptions menampilkan opsi scanning aktif dan meminta konfirmasi
func (s *Service) DisplayScanOptions() (bool, error) {
	print.PrintSubHeader("Opsi Scanning")

	data := [][]string{
		{"Exclude System DB", fmt.Sprintf("%v", s.ScanOptions.ExcludeSystem)},
		{"Include List", fmt.Sprintf("%d database", len(s.ScanOptions.IncludeList))},
		{"Exclude List", fmt.Sprintf("%d database", len(s.ScanOptions.ExcludeList))},
		{"Display Results", fmt.Sprintf("%v", s.ScanOptions.DisplayResults)},
	}

	if s.ScanOptions.Mode == "single" && s.ScanOptions.SourceDatabase != "" {
		data = append(data, []string{"Source Database", s.ScanOptions.SourceDatabase})
	}

	table.Render([]string{"Parameter", "Value"}, data)

	// Konfirmasi
	confirm, err := prompt.Confirm("Apakah Anda ingin melanjutkan?", true)
	if err != nil {
		s.Log.Error("User confirmation error: " + err.Error())
		return false, err
	}

	if !confirm {
		return false, validation.ErrUserCancelled
	}

	s.Log.Info("Proses scanning dilanjutkan.")
	return true, nil
}

// Wrapper methods untuk helper display

func (s *Service) DisplayFilterStats(stats *domain.FilterStats) {
	print.PrintFilterStats(stats, "scan", s.Log)
}

func (s *Service) DisplayScanResult(result *dbscanmodel.ScanResult) {
	helpers.DisplayScanResult(result)
}

func (s *Service) DisplayDetailResults(detailsMap map[string]dbscanmodel.DatabaseDetailInfo) {
	helpers.DisplayDetailResults(detailsMap)
}

func (s *Service) LogDetailResults(detailsMap map[string]dbscanmodel.DatabaseDetailInfo) {
	helpers.LogDetailResults(detailsMap, s.Log)
}
