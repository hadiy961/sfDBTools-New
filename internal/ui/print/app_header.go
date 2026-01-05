// File : internal/ui/print/app_header.go
// Deskripsi : Helper clear screen + app header (judul aplikasi)
// Author : Hadiyatna Muflihun
// Tanggal : 11 November 2025
// Last Modified : 5 Januari 2026

package print

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sfDBTools/pkg/runtimecfg"
	"sfDBTools/pkg/version"

	"github.com/mattn/go-isatty"
)

// ClearScreen clears the terminal screen using platform-specific commands
func ClearScreen() error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		// Linux, macOS, and other Unix-like systems
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return ClearScreenANSI()
	}

	// Only log if there are issues, not for successful operations
	return nil
}

// ClearScreenANSI clears the terminal screen using ANSI escape sequences
func ClearScreenANSI() error {
	// ANSI escape sequence to clear screen and move cursor to top-left
	_, err := fmt.Print("\033[2J\033[H")
	return err
}

// ClearWithMessage clears screen and displays a message
func ClearWithMessage(message string) error {
	if err := ClearScreen(); err != nil {
		return err
	}
	if message != "" {
		fmt.Println(message)
	}
	return nil
}

// ClearAndShowHeader clears screen and shows a formatted header
func ClearAndShowHeader(title string) error {
	if err := ClearScreen(); err != nil {
		return err
	}
	PrintHeader(title)
	return nil
}

// PrintAppHeader menampilkan header aplikasi dan melakukan clear screen bila output TTY.
func PrintAppHeader(title string) {
	if runtimecfg.IsQuiet() || runtimecfg.IsDaemon() {
		return
	}

	fullTitle := fmt.Sprintf("%s v%s - %s", "sfDBTools", version.Version, title)
	// Kalau stdout bukan TTY (misal di-pipe), jangan clear screen.
	if isatty.IsTerminal(os.Stdout.Fd()) {
		_ = ClearAndShowHeader(fullTitle)
		return
	}
	PrintHeader(fullTitle)
}
