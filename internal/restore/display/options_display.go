// File : internal/restore/display/options_display.go
// Deskripsi : Display logic untuk restore options
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-17
// Last Modified : 2025-12-17

package display

import (
	"errors"
	"fmt"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
)

// DisplayConfirmation menampilkan konfirmasi sebelum restore
func DisplayConfirmation(opts map[string]string) error {
	ui.PrintSubHeader("Konfirmasi Restore")
	fmt.Println()

	for key, value := range opts {
		fmt.Printf("  %-20s: %s\n", key, value)
	}

	fmt.Println()

	confirmed, err := input.PromptConfirm("Lanjutkan restore?")
	if err != nil {
		return fmt.Errorf("gagal mendapatkan konfirmasi: %w", err)
	}

	if !confirmed {
		return errors.New("restore dibatalkan oleh user")
	}

	return nil
}
