// File : internal/ui/menu/menu.go
// Deskripsi : Menu interaktif bertingkat berbasis Cobra commands
// Author : Hadiyatna Muflihun
// Tanggal : 23 Januari 2026
// Last Modified : 23 Januari 2026

package menu

import (
	"fmt"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"

	"github.com/spf13/cobra"
)

type itemKind int

const (
	itemCommand itemKind = iota
	itemBack
	itemHome
	itemHelp
	itemExit
)

type menuItem struct {
	Label string
	Kind  itemKind
	Cmd   *cobra.Command
}

// ShowInteractiveMenu menampilkan menu interaktif bertingkat.
// - Root menu menampilkan semua sub-command root.
// - Memilih command yang punya subcommand akan membuka submenu.
// - Memilih command leaf akan menawarkan aksi (jalankan/help) dan bisa kembali.
func ShowInteractiveMenu(root *cobra.Command) error {
	if runtimecfg.IsQuiet() || !isInteractiveTerminal() {
		return showHelpFallback(root)
	}

	stack := []*cobra.Command{root}
	lastBreadcrumb := ""

	for {
		current := stack[len(stack)-1]
		breadcrumb := breadcrumbFromStack(stack)

		// Root header sudah ditangani oleh main.go.
		// Untuk submenu, tampilkan header ketika pindah level.
		if breadcrumb != lastBreadcrumb {
			if len(stack) > 1 {
				print.PrintAppHeader(breadcrumb)
			}
			lastBreadcrumb = breadcrumb
		}

		items := buildMenuItems(current, len(stack) > 1)
		if len(items) == 0 {
			return showHelpFallback(root)
		}

		labels := make([]string, 0, len(items))
		for _, it := range items {
			labels = append(labels, it.Label)
		}

		selected, idx, err := prompt.SelectOne(menuLabelFor(current, len(stack) == 1), labels, -1)
		if err != nil {
			fmt.Println("\nDibatalkan.")
			return nil
		}

		if idx < 0 || idx >= len(items) {
			_ = selected
			continue
		}

		switch items[idx].Kind {
		case itemExit:
			print.PrintInfo("\nTerima kasih telah menggunakan sfdbtools!")
			return nil
		case itemHome:
			stack = stack[:1]
			continue
		case itemBack:
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
			continue
		case itemHelp:
			print.PrintSeparator()
			_ = current.Help()
			prompt.WaitForEnter()
			continue
		case itemCommand:
			cmd := items[idx].Cmd
			if cmd == nil {
				continue
			}

			if hasVisibleSubcommands(cmd) {
				stack = append(stack, cmd)
				continue
			}

			actionLabels := []string{
				"▶️  Jalankan",
				"❓  Lihat help",
				"⬅️  Kembali",
				"🏠  Main Menu",
				"🚪  Keluar",
			}
			action, aidx, aerr := prompt.SelectOne("Aksi:", actionLabels, 0)
			if aerr != nil {
				// Cancel: kembali ke menu yang sama
				_ = action
				continue
			}
			_ = action

			switch aidx {
			case 0: // run
				print.PrintSeparator()
				if err := executeCommand(root, cmd); err != nil {
					print.PrintError(fmt.Sprintf("Gagal menjalankan command: %v", err))
				}
				prompt.WaitForEnter()
				continue
			case 1: // help
				print.PrintSeparator()
				_ = cmd.Help()
				prompt.WaitForEnter()
				continue
			case 2: // back
				if len(stack) > 1 {
					stack = stack[:len(stack)-1]
				}
				continue
			case 3: // home
				stack = stack[:1]
				continue
			case 4: // exit
				print.PrintInfo("\nTerima kasih telah menggunakan sfdbtools!")
				return nil
			default:
				continue
			}
		default:
			continue
		}
	}
}

func showHelpFallback(root *cobra.Command) error {
	fmt.Println("Silakan jalankan 'sfdbtools --help' untuk melihat perintah yang tersedia.")
	print.PrintSeparator()
	return root.Help()
}
