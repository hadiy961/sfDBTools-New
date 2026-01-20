// File : internal/backup/modes/factory.go
// Deskripsi : Factory untuk pembuatan ModeExecutor berdasarkan tipe backup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-11
// Last Modified : 20 Januari 2026

package modes

import (
	"fmt"
	"sfdbtools/internal/shared/consts"
)

// GetExecutor mengembalikan implementasi ModeExecutor yang sesuai berdasarkan mode string.
// State di-pass untuk memungkinkan executor mengakses/update execution state.
func GetExecutor(mode string, svc BackupService, state BackupStateAccessor) (ModeExecutor, error) {
	switch mode {
	case consts.ModeCombined, consts.ModeAll:
		// Combined dan all menggunakan executor yang sama (combined mode)
		// Perbedaannya hanya di nama file output dan header display
		return NewCombinedExecutor(svc, state), nil
	case consts.ModeSingle, consts.ModePrimary, consts.ModeSecondary, consts.ModeSeparated:
		return NewIterativeExecutor(svc, state, mode), nil
	default:
		return nil, fmt.Errorf("mode backup tidak dikenali: %s", mode)
	}
}
