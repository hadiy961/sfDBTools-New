// File : internal/dbscan/dbscan_entry.go
// Deskripsi : Entry point untuk database scan (deprecated, gunakan command.go)
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 16 Desember 2025

package dbscan

import (
	"sfDBTools/internal/types"
)

// ExecuteScanCommand adalah entry point lama (deprecated)
// Gunakan command.ExecuteScanCommand untuk konsistensi dengan package lain
// Deprecated: Use ExecuteScanCommand from command.go
func (s *Service) ExecuteScanCommand(config types.ScanEntryConfig) error {
	return ExecuteScanCommand(s, config)
}
