// File : internal/ui/menu/terminal.go
// Deskripsi : Helper deteksi terminal interaktif
// Author : Hadiyatna Muflihun
// Tanggal : 23 Januari 2026
// Last Modified : 23 Januari 2026

package menu

import (
	"os"

	"github.com/mattn/go-isatty"
)

func isInteractiveTerminal() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
}
