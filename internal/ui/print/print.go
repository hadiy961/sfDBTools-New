// File : internal/ui/print/print.go
// Deskripsi : Print helper (header/info/warn/error) untuk UI facade
// Author : Hadiyatna Muflihun
// Tanggal : 5 Januari 2026
// Last Modified : 5 Januari 2026

package print

import (
	"fmt"
	"sfDBTools/internal/ui/progress"
)

// Println menambah satu baris kosong.
func Println() {
	progress.RunWithSpinnerSuspended(func() {
		fmt.Println()
	})
}
