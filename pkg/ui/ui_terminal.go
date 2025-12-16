package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// GetTerminalSize returns the terminal width and height, with sensible defaults.
func GetTerminalSize() (width, height int, err error) {
	const defaultW, defaultH = 80, 24

	// Try environment variables first
	var w, h int
	if _, err := fmt.Sscanf(os.Getenv("COLUMNS"), "%d", &w); err == nil {
		if _, err := fmt.Sscanf(os.Getenv("LINES"), "%d", &h); err == nil {
			return w, h, nil
		}
	}

	// Try tput on Unix-like systems
	if runtime.GOOS != "windows" {
		if w, h, err := getTputSize(); err == nil {
			return w, h, nil
		}
	}

	return defaultW, defaultH, nil
}

func getTputSize() (width, height int, err error) {
	getDim := func(cmd string) (int, error) {
		out, err := exec.Command("tput", cmd).Output()
		if err != nil {
			return 0, err
		}
		var val int
		_, err = fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &val)
		return val, err
	}

	w, err1 := getDim("cols")
	h, err2 := getDim("lines")
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("tput failed")
	}
	return w, h, nil
}

// PrintBorder prints a horizontal border across the terminal width
func PrintBorder(char rune, width int) {
	if width <= 0 {
		width, _, _ = GetTerminalSize()
		if width <= 0 {
			width = 80 // fallback
		}
	}

	border := strings.Repeat(string(char), width)
	fmt.Println(border)
}

// PrintSeparator prints a separator line
func PrintSeparator() {
	PrintBorder('=', 0)
}

// PrintDashedSeparator prints a dashed separator line
func PrintDashedSeparator() {
	PrintBorder('-', 0)
}

// WaitForEnter waits for the user to press Enter with optional message.
func WaitForEnter(message ...string) {
	msg := "Press Enter to continue..."
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	fmt.Print(msg)
	fmt.Scanln()
}
