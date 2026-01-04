// File : internal/dbscan/display.go
// Deskripsi : Display functions untuk database scanning results
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 2026-01-04

package dbscan

import (
	"fmt"
	"sfDBTools/internal/dbscan/helpers"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

// DisplayScanOptions menampilkan opsi scanning aktif dan meminta konfirmasi
func (s *Service) DisplayScanOptions() (bool, error) {
	ui.PrintSubHeader("Opsi Scanning")

	data := [][]string{
		{"Exclude System DB", fmt.Sprintf("%v", s.ScanOptions.ExcludeSystem)},
		{"Include List", fmt.Sprintf("%d database", len(s.ScanOptions.IncludeList))},
		{"Exclude List", fmt.Sprintf("%d database", len(s.ScanOptions.ExcludeList))},
		{"Display Results", fmt.Sprintf("%v", s.ScanOptions.DisplayResults)},
	}

	if s.ScanOptions.Mode == "single" && s.ScanOptions.SourceDatabase != "" {
		data = append(data, []string{"Source Database", s.ScanOptions.SourceDatabase})
	}

	ui.FormatTable([]string{"Parameter", "Value"}, data)

	// Konfirmasi
	confirm, err := input.AskYesNo("Apakah Anda ingin melanjutkan?", true)
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

// Wrapper methods for pkg/dbscanhelper and pkg/ui display functions

func (s *Service) DisplayFilterStats(stats *types.FilterStats) {
	ui.DisplayFilterStats(stats, "scan", s.Log)
}

func (s *Service) DisplayScanResult(result *types.ScanResult) {
	helpers.DisplayScanResult(result)
}

func (s *Service) DisplayDetailResults(detailsMap map[string]types.DatabaseDetailInfo) {
	helpers.DisplayDetailResults(detailsMap)
}

func (s *Service) LogDetailResults(detailsMap map[string]types.DatabaseDetailInfo) {
	helpers.LogDetailResults(detailsMap, s.Log)
}
