// File : internal/backup/modes/factory.go
// Deskripsi : Factory untuk pembuatan ModeExecutor berdasarkan tipe backup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-11
// Last Modified : 2025-12-11

package modes

import (
	"fmt"
)

// GetExecutor mengembalikan implementasi ModeExecutor yang sesuai berdasarkan mode string
func GetExecutor(mode string, svc BackupService) (ModeExecutor, error) {
	switch mode {
	case "combined":
		return NewCombinedExecutor(svc), nil
	case "single", "primary", "secondary", "separated", "separate":
		return NewIterativeExecutor(svc, mode), nil
	default:
		return nil, fmt.Errorf("mode backup tidak dikenali: %s", mode)
	}
}
