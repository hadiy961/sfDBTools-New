// File : internal/dbscan/command.go
// Deskripsi : Command execution layer untuk database scanning
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 16 Desember 2025

package dbscan

import (
	"sfDBTools/internal/types"
)

// ExecuteScanCommand adalah fungsi wrapper untuk menjalankan scan command
// Pattern ini konsisten dengan backup dan cleanup packages
func ExecuteScanCommand(svc *Service, config types.ScanEntryConfig) error {
	return svc.ExecuteScan(config)
}
