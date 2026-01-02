// File : internal/backup/setup/edit_menu_engine.go
// Deskripsi : Engine menu terpusat untuk edit opsi backup (interactive)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-31
// Last Modified : 2025-12-31

package setup

import (
	"fmt"

	"sfDBTools/pkg/input"
)

type editMenuItem struct {
	Label  string
	Action func() error
}

func (s *Setup) runEditMenuInteractive(items []editMenuItem) error {
	options := make([]string, 0, len(items)+1)
	actions := make(map[string]func() error, len(items))

	for _, it := range items {
		options = append(options, it.Label)
		actions[it.Label] = it.Action
	}
	options = append(options, "Kembali")

	choice, err := input.SelectSingleFromList(options, "Pilih opsi yang ingin diubah")
	if err != nil {
		return fmt.Errorf("gagal memilih opsi untuk diubah: %w", err)
	}
	if choice == "Kembali" {
		return nil
	}

	act := actions[choice]
	if act == nil {
		return fmt.Errorf("opsi tidak dikenali: %s", choice)
	}
	return act()
}
