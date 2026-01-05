// File : internal/app/dbscan/command.go
// Deskripsi : Command execution layer untuk database scanning
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 5 Januari 2026

package dbscan

import (
	dbscanmodel "sfDBTools/internal/app/dbscan/model"
)

// ExecuteScanCommand adalah fungsi wrapper untuk menjalankan scan command
// Pattern ini konsisten dengan backup dan cleanup packages
func ExecuteScanCommand(svc *Service, config dbscanmodel.ScanEntryConfig) error {
	return svc.ExecuteScan(config)
}
