// File : internal/dbscan/dbscan_display.go
// Deskripsi : Display functions untuk database scanning
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 16 Desember 2025

package dbscan

import (
	"fmt"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/dbscanhelper"
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
		s.Log.Error("gagal mendapatkan konfirmasi user: " + askErr.Error())
		return false, askErr
	}
	if !confirm {
		return false, types.ErrUserCancelled
	}
	s.Log.Info("Proses scanning dilanjutkan.")
	return true, nil
}

// DisplayFilterStats menampilkan statistik hasil pemfilteran database.
func (s *Service) DisplayFilterStats(stats *types.DatabaseFilterStats) {
	ui.DisplayFilterStats(stats, "scan", s.Log)
}

// DisplayScanResult menampilkan hasil scanning (wrapper untuk helper)
func (s *Service) DisplayScanResult(result *types.ScanResult) {
	dbscanhelper.DisplayScanResult(result)
}

// DisplayDetailResults menampilkan detail hasil scanning (wrapper untuk helper)
func (s *Service) DisplayDetailResults(detailsMap map[string]types.DatabaseDetailInfo) {
	dbscanhelper.DisplayDetailResults(detailsMap)
}

// LogDetailResults menulis detail hasil scanning ke logger (wrapper untuk helper)
func (s *Service) LogDetailResults(detailsMap map[string]types.DatabaseDetailInfo) {
	dbscanhelper.LogDetailResults(detailsMap, s.Log)
}
