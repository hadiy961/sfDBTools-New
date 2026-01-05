// File : internal/app/dbscan/helpers/display.go
// Deskripsi : Helper functions untuk menampilkan hasil scanning (general purpose)
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 5 Januari 2026
package helpers

import (
	"fmt"

	dbscanmodel "sfDBTools/internal/app/dbscan/model"
	applog "sfDBTools/internal/services/log"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/ui"
)

// DisplayScanResult menampilkan hasil scanning ke UI
func DisplayScanResult(result *dbscanmodel.ScanResult) {
	ui.PrintHeader("HASIL SCANNING")

	data := [][]string{
		{"Total Database", fmt.Sprintf("%d", result.TotalDatabases)},
		{"Berhasil", ui.ColorText(fmt.Sprintf("%d", result.SuccessCount), consts.UIColorGreen)},
		{"Gagal", ui.ColorText(fmt.Sprintf("%d", result.FailedCount), consts.UIColorRed)},
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

// DisplayDetailResults menampilkan detail hasil scanning ke UI
func DisplayDetailResults(detailsMap map[string]dbscanmodel.DatabaseDetailInfo) {
	ui.PrintHeader("DETAIL HASIL SCANNING")

	headers := []string{"Database", "Size", "Tables", "Procedures", "Functions", "Views", "Grants", "Status"}
	var rows [][]string

	for _, detail := range detailsMap {
		status := ui.ColorText("✓ OK", consts.UIColorGreen)
		if detail.Error != "" {
			status = ui.ColorText("✗ Error", consts.UIColorRed)
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

// LogScanResult menulis hasil scanning ke logger (untuk background mode)
func LogScanResult(result *dbscanmodel.ScanResult, logger applog.Logger, scanID string) {
	logger.Infof("[%s] ========================================", scanID)
	logger.Infof("[%s] HASIL SCANNING", scanID)
	logger.Infof("[%s] ========================================", scanID)
	logger.Infof("[%s] Total Database  : %d", scanID, result.TotalDatabases)
	logger.Infof("[%s] Berhasil        : %d", scanID, result.SuccessCount)
	logger.Infof("[%s] Gagal           : %d", scanID, result.FailedCount)
	logger.Infof("[%s] Durasi          : %s", scanID, result.Duration)

	if len(result.Errors) > 0 {
		logger.Warnf("[%s] Terdapat %d error saat scanning:", scanID, len(result.Errors))
		for i, errMsg := range result.Errors {
			logger.Warnf("[%s]   %d. %s", scanID, i+1, errMsg)
		}
	}
}

// LogDetailResults menulis detail hasil scanning ke logger (untuk background mode)
func LogDetailResults(detailsMap map[string]dbscanmodel.DatabaseDetailInfo, logger applog.Logger) {
	logger.Info("=== Detail Hasil Scanning ===")

	for dbName, detail := range detailsMap {
		if detail.Error != "" {
			logger.Warnf("Database: %s - Status: ERROR - %s", dbName, detail.Error)
		}
	}
}
