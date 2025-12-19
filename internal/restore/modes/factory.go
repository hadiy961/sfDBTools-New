// File : internal/restore/modes/factory.go
// Deskripsi : Factory untuk pembuatan RestoreExecutor berdasarkan tipe restore
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package modes

import (
	"fmt"
)

// GetExecutor mengembalikan implementasi RestoreExecutor yang sesuai berdasarkan mode string
func GetExecutor(mode string, svc RestoreService) (RestoreExecutor, error) {
	switch mode {
	case "single":
		return NewSingleExecutor(svc), nil
	case "primary":
		return NewPrimaryExecutor(svc), nil
	case "all":
		return NewAllExecutor(svc), nil
	case "selection":
		return NewSelectionExecutor(svc), nil
	default:
		return nil, fmt.Errorf("mode restore tidak dikenali: %s", mode)
	}
}
