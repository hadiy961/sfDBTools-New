package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	app_log "sfDBTools/internal/applog"
	"strings"
)

// ConvertSliceTo2D
func ConvertSliceTo2D(slice []string) [][]string {
	result := make([][]string, len(slice))
	for i, v := range slice {
		result[i] = []string{v}
	}
	return result
}

// GetTerminalSize returns the terminal width and height
func GetTerminalSize() (width, height int, err error) {
	lg := app_log.NewLogger()

	// First try using environment variables (more reliable)
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if lines := os.Getenv("LINES"); lines != "" {
			var w, h int
			if _, err := fmt.Sscanf(cols, "%d", &w); err == nil {
				if _, err := fmt.Sscanf(lines, "%d", &h); err == nil {
					// Only log terminal size detection failures, not successes for cleaner output
					return w, h, nil
				}
			}
		}
	}

	// Try using tput command (more reliable than stty)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// On Windows, try powershell
		cmd = exec.Command("powershell", "-Command", "(Get-Host).UI.RawUI.WindowSize")
	default:
		// Try tput first (more reliable)
		if width, height, err := getTputSize(); err == nil {
			return width, height, nil
		}

		// Fallback to stty
		cmd = exec.Command("stty", "size")
	}

	if cmd == nil {
		lg.Debug("No terminal size command available, using defaults")
		return 80, 24, nil
	}

	output, err := cmd.Output()
	if err != nil {
		lg.Debug("Failed to get terminal size via command, using defaults ", lg.WithError(err))
		// Return default values without error
		return 80, 24, nil
	}

	outputStr := strings.TrimSpace(string(output))

	if runtime.GOOS == "windows" {
		// Parse Windows PowerShell output (simplified)
		lg.Debug("Windows terminal size detection using defaults")
		return 80, 24, nil
	} else {
		// Parse Unix stty output: "height width"
		var h, w int
		if _, err := fmt.Sscanf(outputStr, "%d %d", &h, &w); err != nil {
			lg.Debug("Failed to parse terminal size, using defaults ", lg.WithError(err))
			return 80, 24, nil
		}
		lg.Debug("Terminal size detected via stty ", lg.WithField("width", w), lg.WithField("height", h))
		return w, h, nil
	}
}

// getTputSize tries to get terminal size using tput command
func getTputSize() (width, height int, err error) {
	// Get columns
	colsCmd := exec.Command("tput", "cols")
	colsOutput, err := colsCmd.Output()
	if err != nil {
		return 0, 0, err
	}

	// Get lines
	linesCmd := exec.Command("tput", "lines")
	linesOutput, err := linesCmd.Output()
	if err != nil {
		return 0, 0, err
	}

	var w, h int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(colsOutput)), "%d", &w); err != nil {
		return 0, 0, err
	}

	if _, err := fmt.Sscanf(strings.TrimSpace(string(linesOutput)), "%d", &h); err != nil {
		return 0, 0, err
	}

	// Remove debug logging for terminal size detection success for cleaner output
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

// WaitForEnter waits for the user to press Enter
func WaitForEnter() {
	fmt.Print("Press Enter to continue...")
	fmt.Scanln()
}

// WaitForEnterWithMessage waits for the user to press Enter with a custom message
func WaitForEnterWithMessage(message string) {
	fmt.Print(message)
	fmt.Scanln()
}
