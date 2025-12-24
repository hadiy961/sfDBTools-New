// File : internal/restore/modes/factory.go
// Deskripsi : Factory untuk pembuatan RestoreExecutor berdasarkan tipe restore
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package modes

import (
	"fmt"
	"sfDBTools/pkg/consts"
)

// GetExecutor mengembalikan implementasi RestoreExecutor yang sesuai berdasarkan mode string
func GetExecutor(mode string, svc RestoreService) (RestoreExecutor, error) {
	switch mode {
	case consts.ModeSingle:
		return NewSingleExecutor(svc), nil
	case consts.ModePrimary:
		return NewPrimaryExecutor(svc), nil
	case consts.ModeSecondary:
		return NewSecondaryExecutor(svc), nil
	case consts.ModeAll:
		return NewAllExecutor(svc), nil
	case consts.ModeSelection:
		return NewSelectionExecutor(svc), nil
	case consts.ModeCustom:
		return NewCustomExecutor(svc), nil
	default:
		return nil, fmt.Errorf("mode restore tidak dikenali: %s", mode)
	}
}
