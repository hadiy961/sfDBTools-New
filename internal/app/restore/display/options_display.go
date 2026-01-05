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
	"sort"
)

// DisplayConfirmation menampilkan konfirmasi sebelum restore
func DisplayConfirmation(opts map[string]string) error {
	ui.PrintSubHeader("Konfirmasi Restore")
	fmt.Println()

	keys := make([]string, 0, len(opts))
	for k := range opts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	rows := make([][]string, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, []string{k, opts[k]})
	}

	ui.FormatTable([]string{"Parameter", "Value"}, rows)
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
