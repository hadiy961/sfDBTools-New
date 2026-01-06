// File : internal/restore/display/options_display.go
// Deskripsi : Display logic untuk restore options
// Author : Hadiyatna Muflihun
// Tanggal : 17 Desember 2025
// Last Modified : 17 Desember 2025

package display

import (
	"errors"
	"fmt"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/internal/ui/table"
	"sort"
)

// DisplayConfirmation menampilkan konfirmasi sebelum restore
func DisplayConfirmation(opts map[string]string) error {
	print.PrintSubHeader("Konfirmasi Restore")
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

	table.Render([]string{"Parameter", "Value"}, rows)
	fmt.Println()

	confirmed, err := prompt.PromptConfirm("Lanjutkan restore?")
	if err != nil {
		return fmt.Errorf("gagal mendapatkan konfirmasi: %w", err)
	}

	if !confirmed {
		return errors.New("restore dibatalkan oleh user")
	}

	return nil
}
